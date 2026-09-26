import assert from 'node:assert/strict';
import {spawnSync} from 'node:child_process';
import {chmodSync, mkdirSync, mkdtempSync, rmSync, symlinkSync} from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import {fileURLToPath} from 'node:url';

import {fixtureEnvironment, git, write} from './git-fixture.mjs';

const cli = fileURLToPath(new URL('./size.mjs', import.meta.url));
const policyPath = 'docs/engineering/coding-standards.md';
const emptyPolicy = '## Inherited size baseline\n\n| File | Adoption ceiling (physical lines) |\n|---|---:|\n\n## Construction and readability\n';

function createRepository(t, {withPolicy = true} = {}) {
	const root = mkdtempSync('/tmp/engineering-size-');
	t.after(() => rmSync(root, {recursive: true, force: true}));
	git(root, 'init', '-b', 'main');
	git(root, 'config', 'user.email', 'fixture@example.test');
	git(root, 'config', 'user.name', 'Fixture');
	if (withPolicy) {
		write(root, policyPath, emptyPolicy);
	}
	git(root, 'add', '.');
	git(root, 'commit', '--allow-empty', '-m', 'Fixture base');
	return root;
}

function check(root, ...args) {
	return spawnSync(process.execPath, [cli, ...(args.length ? args : ['--base', 'HEAD'])], {cwd: root, encoding: 'utf8', env: fixtureEnvironment()});
}

test('rejects a tracked source with 501 physical lines and a space in its path', t => {
	const root = createRepository(t);
	write(root, 'src/too large.ts', 'statement\n'.repeat(501));
	git(root, 'add', '.');
	const result = check(root);
	assert.equal(result.status, 1);
	assert.match(result.stderr, /src\/too large\.ts: 501 lines exceeds 500/);
});

test('counts an unterminated last line in non-ignored new sources', t => {
	const root = createRepository(t);
	write(root, 'src/new\nsource.ts', 'statement\r\n'.repeat(500) + 'last');
	const result = check(root);
	assert.equal(result.status, 1);
	assert.ok(result.stderr.includes('src/new\nsource.ts: 501 lines exceeds 500'));
});

test('accepts empty and exactly 500-line files without duplicate inventory entries', t => {
	const root = createRepository(t);
	write(root, 'empty.ts', '');
	write(root, 'limit.ts', 'line\r\n'.repeat(500));
	write(root, 'unterminated.ts', 'line\n'.repeat(499) + 'last');
	git(root, 'add', '.');
	const result = check(root);
	assert.equal(result.status, 0, result.stderr);
	assert.match(result.stdout, /size: checked 3 files/);
});

test('includes handwritten languages, declarations, executable fixtures and build instructions', t => {
	const root = createRepository(t);
	const sources = [
		'backend/code.go', 'types.d.ts', 'component.tsx', 'module.mjs', 'style.css',
		'migration.sql', 'script.sh', 'new-language.zig', 'Dockerfile', 'backend/Makefile',
		'.github/workflows/check.yml', 'test/fixtures/helper.js', 'test/fixtures/executable.txt'
	];
	for (const name of sources) {
		write(root, name, (name.endsWith('.txt') ? '#!/bin/sh\n' : 'source\n') + 'line\n'.repeat(500));
	}
	const result = check(root);
	assert.equal(result.status, 1);
	for (const name of sources) {
		assert.ok(result.stderr.includes(`${name}: 501 lines exceeds 500`), name);
	}
});

test('excludes named generated paths, ignored new files, documentation, manifests and inert data', t => {
	const root = createRepository(t);
	const excluded = [
		'node_modules/vendor.ts', '.next/types/generated.ts', 'out/bundle.js', 'coverage/report.js',
		'.gitnexus/generated.ts', 'next-env.d.ts', 'README.md', 'manual.rst', 'package.json',
		'backend/go.mod', 'backend/go.sum', 'bun.lock', 'test/fixtures/response.json',
		'test/fixtures/response.yaml', 'public/photo.jpg'
	];
	for (const name of excluded) {
		write(root, name, 'data\n'.repeat(501));
	}
	write(root, '.gitignore', 'ignored.ts\n');
	write(root, 'ignored.ts', 'line\n'.repeat(501));
	write(root, 'binary.asset', Buffer.from([0, ...Buffer.from('line\n'.repeat(501))]));
	git(root, 'add', '.');
	const result = check(root);
	assert.equal(result.status, 0, result.stderr);
});

test('excludes removed tracked files from current reads', t => {
	const root = createRepository(t);
	write(root, 'removed.ts', 'line\n'.repeat(501));
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Before deletion');
	rmSync(path.join(root, 'removed.ts'));
	const result = check(root);
	assert.equal(result.status, 0, result.stderr);
});

test('reports unreadable source inputs as actionable failures', t => {
	const root = createRepository(t);
	write(root, 'unreadable.ts', 'line\n');
	git(root, 'add', '.');
	rmSync(path.join(root, 'unreadable.ts'));
	mkdirSync(path.join(root, 'unreadable.ts'));
	const result = check(root);
	assert.equal(result.status, 1);
	assert.match(result.stderr, /Cannot read source unreadable\.ts/);
	assert.doesNotMatch(result.stderr, /at ModuleJob/);
});

test('ignores directory symlinks while checking linked source files', t => {
	const root = createRepository(t);
	write(root, '.agents/skills/example/SKILL.md', '# Skill\n');
	mkdirSync(path.join(root, '.claude/skills'), {recursive: true});
	symlinkSync('../../.agents/skills/example', path.join(root, '.claude/skills/example'));
	const directories = check(root);
	assert.equal(directories.status, 0, directories.stderr);
	write(root, 'linked-source.txt', 'line\n'.repeat(501));
	symlinkSync('linked-source.txt', path.join(root, 'source.ts'));
	const source = check(root);
	assert.equal(source.status, 1);
	assert.match(source.stderr, /source\.ts: 501 lines exceeds 500/);
});

test('fails an invalid change base without evaluating it as shell input', t => {
	const root = createRepository(t);
	const result = check(root, '--base', 'missing; touch unexpected.ts');
	assert.equal(result.status, 1);
	assert.match(result.stderr, /Invalid change base/);
	assert.equal(git(root, 'status', '--porcelain'), '');
});

test('defaults to the main merge base and requires an explicit base when main is unavailable', t => {
	const root = createRepository(t);
	const base = git(root, 'rev-parse', 'HEAD');
	git(root, 'checkout', '-b', 'feature');
	write(root, 'small.ts', 'line\n');
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Feature');
	const result = spawnSync(process.execPath, [cli], {cwd: root, encoding: 'utf8', env: fixtureEnvironment()});
	assert.equal(result.status, 0, result.stderr);
	assert.ok(result.stdout.includes(base));
	git(root, 'branch', '-D', 'main');
	const missing = spawnSync(process.execPath, [cli], {cwd: root, encoding: 'utf8', env: fixtureEnvironment()});
	assert.equal(missing.status, 1);
	assert.match(missing.stderr, /Cannot resolve main merge base; supply --base/);
});

test('accepts inherited oversized files within their documented ceiling and base count', t => {
	const root = createRepository(t);
	write(root, policyPath, emptyPolicy.replace('|---|---:|', '|---|---:|\n| `legacy.go` | 600 |'));
	write(root, 'legacy.go', 'line\n'.repeat(600));
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Inherited baseline');
	const result = check(root);
	assert.equal(result.status, 0, result.stderr);
	assert.match(result.stdout, /size: checked 1 files/);
});

test('rejects missing, malformed, duplicate and invalid inherited baseline records', t => {
	const root = createRepository(t);
	const malformed = [
		'no baseline table',
		emptyPolicy.replace('|---|---:|', '|invalid|separator|'),
		emptyPolicy.replace('|---|---:|', '|---|---:|\n| `legacy.go` | many |'),
		emptyPolicy.replace('|---|---:|', '|---|---:|\n| `legacy.go` | 600 |\n| `legacy.go` | 600 |'),
		emptyPolicy.replace('|---|---:|', '|---|---:|\n| `legacy.go` | 500 |'),
		emptyPolicy.replace('|---|---:|', '|---|---:|\n| `../legacy.go` | 600 |'),
		emptyPolicy.replace('|---|---:|', '|---|---:|\n| `src/*.go` | 600 |'),
		emptyPolicy.replace('|---|---:|', '|---|---:|\n| `legacy.go` | 9007199254740992 |'),
		emptyPolicy + emptyPolicy
	];
	for (const policy of malformed) {
		write(root, policyPath, policy);
		const result = check(root);
		assert.equal(result.status, 1, policy);
		assert.match(result.stderr, /Invalid inherited size baseline/);
	}
});

test('rejects inherited growth relative to the supplied base even below the documented ceiling', t => {
	const root = createRepository(t);
	write(root, policyPath, emptyPolicy.replace('|---|---:|', '|---|---:|\n| `legacy.go` | 600 |'));
	write(root, 'legacy.go', 'line\n'.repeat(550));
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Smaller change base');
	const smallerBase = git(root, 'rev-parse', 'HEAD');
	write(root, 'legacy.go', 'line\n'.repeat(560));
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Later revision');
	assert.equal(check(root).status, 0);
	const result = check(root, '--base', smallerBase);
	assert.equal(result.status, 1);
	assert.match(result.stderr, /legacy\.go: 560 lines exceeds change base 550/);
});

test('rejects raising an inherited ceiling and exceeding the documented ceiling', t => {
	const root = createRepository(t);
	const policy = emptyPolicy.replace('|---|---:|', '|---|---:|\n| `legacy.go` | 600 |');
	write(root, policyPath, policy);
	write(root, 'legacy.go', 'line\n'.repeat(600));
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Inherited baseline');
	write(root, policyPath, policy.replace('600', '601'));
	const raised = check(root);
	assert.equal(raised.status, 1);
	assert.match(raised.stderr, /legacy\.go: ceiling 601 exceeds inherited ceiling 600/);
	write(root, policyPath, policy);
	write(root, 'legacy.go', 'line\n'.repeat(601));
	const grown = check(root);
	assert.equal(grown.status, 1);
	assert.match(grown.stderr, /legacy\.go: 601 lines exceeds 600/);
});

test('rejects a new file given an inherited exception without base provenance', t => {
	const root = createRepository(t);
	write(root, 'new.go', 'line\n'.repeat(600));
	write(root, policyPath, emptyPolicy.replace('|---|---:|', '|---|---:|\n| `new.go` | 600 |'));
	const result = check(root);
	assert.equal(result.status, 1);
	assert.match(result.stderr, /new\.go: exception has no inherited source at the change base/);
});

test('requires retirement when an exception is resolved or its file is removed', t => {
	const root = createRepository(t);
	write(root, policyPath, emptyPolicy.replace('|---|---:|', '|---|---:|\n| `legacy.go` | 600 |'));
	write(root, 'legacy.go', 'line\n'.repeat(600));
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Inherited baseline');
	write(root, 'legacy.go', 'line\n'.repeat(500));
	const resolved = check(root);
	assert.equal(resolved.status, 1);
	assert.match(resolved.stderr, /legacy\.go: retire the resolved exception/);
	rmSync(path.join(root, 'legacy.go'));
	const removed = check(root);
	assert.equal(removed.status, 1);
	assert.match(removed.stderr, /legacy\.go: retire the resolved exception/);
	write(root, policyPath, emptyPolicy);
	assert.equal(check(root).status, 0);
});

test('requires a lower documented ceiling after extracting an oversized file', t => {
	const root = createRepository(t);
	const policy = emptyPolicy.replace('|---|---:|', '|---|---:|\n| `legacy.go` | 600 |');
	write(root, policyPath, policy);
	write(root, 'legacy.go', 'line\n'.repeat(600));
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Inherited baseline');
	write(root, 'legacy.go', 'line\n'.repeat(550));
	const stale = check(root);
	assert.equal(stale.status, 1);
	assert.match(stale.stderr, /legacy\.go: lower the ceiling to 550 after extraction/);
	write(root, policyPath, policy.replace('600', '550'));
	assert.equal(check(root).status, 0);
});

test('accepts a Git rename only when its reviewed exception is explicitly carried without raising the ceiling', t => {
	const root = createRepository(t);
	const policy = emptyPolicy.replace('|---|---:|', '|---|---:|\n| `legacy.go` | 600 |');
	write(root, policyPath, policy);
	write(root, 'legacy.go', 'line\n'.repeat(600));
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Inherited baseline');
	git(root, 'mv', 'legacy.go', 'renamed.go');
	const implicit = check(root);
	assert.equal(implicit.status, 1);
	assert.match(implicit.stderr, /renamed\.go: 600 lines exceeds 500/);
	write(root, policyPath, policy.replace('legacy.go', 'renamed.go'));
	const explicit = check(root);
	assert.equal(explicit.status, 0, explicit.stderr);
	write(root, policyPath, policy.replace('legacy.go', 'renamed.go').replace('600', '601'));
	const raised = check(root);
	assert.equal(raised.status, 1);
	assert.match(raised.stderr, /renamed\.go: ceiling 601 exceeds inherited ceiling 600/);
});

test('rejects missing or unknown CLI arguments and provides focused usage', t => {
	const root = createRepository(t);
	for (const args of [['--base'], ['--unknown']]) {
		const result = spawnSync(process.execPath, [cli, ...args], {cwd: root, encoding: 'utf8', env: fixtureEnvironment()});
		assert.equal(result.status, 1);
		assert.match(result.stderr, /Usage: node scripts\/checks\/size\.mjs \[--base <revision>\]/);
	}
	const help = spawnSync(process.execPath, [cli, '--help'], {cwd: root, encoding: 'utf8', env: fixtureEnvironment()});
	assert.equal(help.status, 0);
	assert.match(help.stdout, /Usage:/);
});

test('rejects repeated comparison bases before evaluating size inputs', t => {
	const root = createRepository(t);
	for (const revisions of [['missing-review-base', 'HEAD'], ['HEAD', 'missing-review-base'], ['HEAD', 'HEAD']]) {
		const result = check(root, '--base', revisions[0], '--base', revisions[1]);
		assert.equal(result.status, 1, `Repeated bases must fail: ${revisions.join(', ')}`);
		assert.match(result.stderr, /Usage: node scripts\/checks\/size\.mjs \[--base <revision>\]/);
		assert.doesNotMatch(result.stdout, /size: checked/);
	}
});

test('excludes only the named Go coverage profile rather than all out files', t => {
	const root = createRepository(t);
	write(root, 'backend/coverage.out', 'mode: atomic\n' + 'coverage data\n'.repeat(600));
	write(root, 'handwritten.out', 'source\n'.repeat(501));
	const result = check(root);
	assert.equal(result.status, 1);
	assert.match(result.stderr, /handwritten\.out: 501 lines exceeds 500/);
	assert.doesNotMatch(result.stderr, /backend\/coverage\.out/);
});

test('excludes executable binary assets while retaining executable data-named fixtures', t => {
	const root = createRepository(t);
	write(root, 'binary.asset', Buffer.from([0, ...Buffer.from('data\n'.repeat(501))]));
	write(root, 'test/fixtures/executable.json', 'statement\n'.repeat(501));
	chmodSync(path.join(root, 'binary.asset'), 0o755);
	chmodSync(path.join(root, 'test/fixtures/executable.json'), 0o755);
	const result = check(root);
	assert.equal(result.status, 1);
	assert.match(result.stderr, /test\/fixtures\/executable\.json: 501 lines exceeds 500/);
	assert.doesNotMatch(result.stderr, /binary\.asset/);
});

test('proves initial policy exceptions against oversized source at the base including explicit renames', t => {
	const root = createRepository(t, {withPolicy: false});
	write(root, 'legacy.go', 'line\n'.repeat(600));
	write(root, 'small.go', 'line\n'.repeat(500));
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Before engineering policy');
	const policy = emptyPolicy.replace('|---|---:|', '|---|---:|\n| `legacy.go` | 600 |');
	write(root, policyPath, policy);
	assert.equal(check(root).status, 0);
	git(root, 'mv', 'legacy.go', 'renamed.go');
	write(root, policyPath, policy.replace('legacy.go', 'renamed.go'));
	assert.equal(check(root).status, 0);
	write(root, policyPath, policy.replace('legacy.go', 'renamed.go').replace('600', '601'));
	assert.match(check(root).stderr, /renamed\.go: ceiling 601 exceeds inherited ceiling 600/);
	write(root, policyPath, policy.replace('legacy.go', 'renamed.go'));
	write(root, 'renamed.go', 'line\n'.repeat(601));
	assert.match(check(root).stderr, /renamed\.go: 601 lines exceeds change base 600/);
	write(root, 'new.go', 'line\n'.repeat(600));
	write(root, policyPath, policy.replace('legacy.go', 'new.go'));
	assert.match(check(root).stderr, /new\.go: exception has no inherited source/);
	write(root, 'small.go', 'line\n'.repeat(600));
	write(root, policyPath, policy.replace('legacy.go', 'small.go'));
	assert.match(check(root).stderr, /small\.go: exception has no inherited source/);
});

test('refuses initial-adoption fallback after a historical policy was deleted', t => {
	const root = createRepository(t);
	write(root, 'legacy.go', 'line\n'.repeat(600));
	rmSync(path.join(root, policyPath));
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Policy removed');
	write(root, policyPath, emptyPolicy.replace('|---|---:|', '|---|---:|\n| `legacy.go` | 600 |'));
	const result = check(root);
	assert.equal(result.status, 1);
	assert.match(result.stderr, /Historical size policy was deleted/);
});

test('identifies malformed and unreadable historical policies without using adoption fallback', t => {
	const root = createRepository(t);
	write(root, 'legacy.go', 'line\n'.repeat(600));
	write(root, policyPath, 'broken historical table');
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Malformed historical policy');
	write(root, policyPath, emptyPolicy.replace('|---|---:|', '|---|---:|\n| `legacy.go` | 600 |'));
	const malformed = check(root);
	assert.equal(malformed.status, 1);
	assert.match(malformed.stderr, /Cannot validate historical size policy .*coding-standards\.md.*Invalid inherited size baseline/);
	const blob = git(root, 'rev-parse', `HEAD:${policyPath}`);
	rmSync(path.join(root, '.git', 'objects', blob.slice(0, 2), blob.slice(2)));
	const unreadable = check(root);
	assert.equal(unreadable.status, 1);
	assert.match(unreadable.stderr, /Cannot validate historical size policy .*coding-standards\.md/);
});

test('fails initial-adoption provenance when the base history is shallow', t => {
	const original = createRepository(t);
	write(original, 'legacy.go', 'line\n'.repeat(600));
	rmSync(path.join(original, policyPath));
	git(original, 'add', '.');
	git(original, 'commit', '-m', 'Policy deletion beyond shallow boundary');
	const root = mkdtempSync('/tmp/engineering-size-shallow-');
	t.after(() => rmSync(root, {recursive: true, force: true}));
	git(original, 'clone', '--depth=1', `file://${original}`, root);
	write(root, policyPath, emptyPolicy.replace('|---|---:|', '|---|---:|\n| `legacy.go` | 600 |'));
	const result = check(root);
	assert.equal(result.status, 1);
	assert.match(result.stderr, /Cannot establish size policy history.*shallow/);
});

test('checks valid JavaScript containing NUL inside a comment', t => {
	const root = createRepository(t);
	write(root, 'nul-comment.js', '// NUL is valid here: \0\n' + '// source\n'.repeat(500));
	const syntax = spawnSync(process.execPath, ['--check', 'nul-comment.js'], {cwd: root, encoding: 'utf8', env: fixtureEnvironment()});
	assert.equal(syntax.status, 0, syntax.stderr);
	const result = check(root);
	assert.equal(result.status, 1);
	assert.match(result.stderr, /nul-comment\.js: 501 lines exceeds 500/);
});

test('excludes non-executable GPX fixture data while retaining executable GPX-named scripts', t => {
	const root = createRepository(t);
	const name = 'test/fixtures/track.gpx';
	write(root, name, '<?xml version="1.0"?>\n<gpx>\n' + '<wpt lat="1" lon="2" />\n'.repeat(500) + '</gpx>\n');
	const data = check(root);
	assert.equal(data.status, 0, data.stderr);
	write(root, name, '#!/bin/sh\n' + 'echo coordinate\n'.repeat(500));
	const executable = check(root);
	assert.equal(executable.status, 1);
	assert.match(executable.stderr, /test\/fixtures\/track\.gpx: 501 lines exceeds 500/);
});

test('preserves base executable mode when proving initial fixture exceptions', t => {
	const root = createRepository(t, {withPolicy: false});
	const executable = 'test/fixtures/executable script.txt';
	const data = 'test/fixtures/data.txt';
	write(root, executable, 'echo coordinate\n'.repeat(600));
	write(root, data, 'coordinate\n'.repeat(600));
	chmodSync(path.join(root, executable), 0o755);
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Executable fixture before policy');
	const policy = emptyPolicy.replace('|---|---:|', `|---|---:|\n| \`${executable}\` | 600 |`);
	write(root, policyPath, policy);
	const inherited = check(root);
	assert.equal(inherited.status, 0, inherited.stderr);
	chmodSync(path.join(root, data), 0o755);
	write(root, policyPath, policy.replace('|---|---:|', `|---|---:|\n| \`${data}\` | 600 |`));
	const newlyExecutable = check(root);
	assert.equal(newlyExecutable.status, 1);
	assert.match(newlyExecutable.stderr, /test\/fixtures\/data\.txt: exception has no inherited source/);
});

test('isolates fixture subprocesses from inherited Git configuration and repository overrides', t => {
	const settings = mkdtempSync('/tmp/engineering-size-git-config-');
	t.after(() => rmSync(settings, {recursive: true, force: true}));
	write(settings, 'global.conf', '[commit]\n gpgSign = true\n[gpg]\n program = /missing/fixture-signing-program\n');
	write(settings, 'system.conf', '[core]\n bare = true\n');
	const inheritedEnvironment = process.env;
	try {
		process.env = {...inheritedEnvironment, GIT_CONFIG_GLOBAL: path.join(settings, 'global.conf')};
		assert.equal(check(createRepository(t)).status, 0);
		process.env = {
			...process.env,
			GIT_CONFIG_SYSTEM: path.join(settings, 'system.conf'),
			GIT_CONFIG_NOSYSTEM: '0',
			GIT_CONFIG_COUNT: '1',
			GIT_CONFIG_KEY_0: 'core.bare',
			GIT_CONFIG_VALUE_0: 'true',
			GIT_DIR: path.join(settings, 'missing.git'),
			GIT_WORK_TREE: settings,
			GIT_INDEX_FILE: path.join(settings, 'user-index'),
			GIT_OBJECT_DIRECTORY: path.join(settings, 'user-objects')
		};
		const result = check(createRepository(t));
		assert.equal(result.status, 0, result.stderr);
	} finally {
		process.env = inheritedEnvironment;
	}
});
