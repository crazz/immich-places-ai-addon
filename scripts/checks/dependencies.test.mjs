import assert from 'node:assert/strict';
import {spawnSync} from 'node:child_process';
import {mkdirSync, rmSync} from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import {fileURLToPath} from 'node:url';

import {check, createRepository, git, write} from './dependency-fixture.mjs';

const cli = fileURLToPath(new URL('./dependencies.mjs', import.meta.url));

test('provides focused dependency usage and rejects invalid arguments', () => {
	const help = spawnSync(process.execPath, [cli, '--help'], {encoding: 'utf8'});
	assert.equal(help.status, 0, 'dependency help command must succeed');
	assert.match(help.stdout, /Usage: node scripts\/checks\/dependencies\.mjs \[--base <revision>\]/);
	for (const args of [['--base'], ['--unknown']]) {
		const result = spawnSync(process.execPath, [cli, ...args], {encoding: 'utf8'});
		assert.equal(result.status, 1);
		assert.match(result.stderr, /Usage:/);
	}
});

test('rejects repeated comparison bases before evaluating dependency inputs', t => {
	const root = createRepository(t);
	for (const revisions of [['missing-review-base', 'HEAD'], ['HEAD', 'missing-review-base'], ['HEAD', 'HEAD']]) {
		const result = check(root, '--base', revisions[0], '--base', revisions[1]);
		assert.equal(result.status, 1, `Repeated bases must fail: ${revisions.join(', ')}`);
		assert.match(result.stderr, /Usage: node scripts\/checks\/dependencies\.mjs \[--base <revision>\]/);
		assert.doesNotMatch(result.stdout, /dependencies: checked/);
	}
});

test('resolves aliases and re-exports while leaving external packages outside the graph', t => {
	const root = createRepository(t);
	write(root, 'src/entry point.ts', "import {value} from '@/barrel'; import React from 'react'; export {value, React};\n");
	write(root, 'src/barrel.ts', "export {value} from './value.js';\n");
	write(root, 'src/value.ts', 'export const value = 1;\n');
	const result = check(root);
	assert.equal(result.status, 0, result.stderr);
	assert.match(result.stdout, /dependencies: checked 4 frontend files, 2 runtime edges/);
	assert.ok(result.stdout.includes(git(root, 'rev-parse', 'HEAD')));
});

test('rejects unresolved local imports including aliases with actionable source paths', t => {
	const root = createRepository(t);
	for (const specifier of ['./missing', '@/missing']) {
		write(root, 'src/consumer.ts', `import '${specifier}';\n`);
		const result = check(root);
		assert.equal(result.status, 1, specifier);
		assert.ok(result.stderr.includes(`src/consumer.ts: unresolved local import ${specifier}`), result.stderr);
	}
});


test('fails closed on TypeScript parse, read and configuration errors', t => {
	const root = createRepository(t);
	write(root, 'src/broken.ts', 'export const = ;\n');
	let result = check(root);
	assert.equal(result.status, 1);
	assert.match(result.stderr, /src\/broken\.ts.*TypeScript parse error/);
	git(root, 'add', '.');
	rmSync(path.join(root, 'src/broken.ts'));
	mkdirSync(path.join(root, 'src/broken.ts'));
	result = check(root);
	assert.equal(result.status, 1);
	assert.match(result.stderr, /Cannot read source src\/broken\.ts/);
	rmSync(path.join(root, 'src/broken.ts'), {recursive: true});
	for (const config of ['{invalid', '{"compilerOptions":{"moduleResolution":"impossible"}}', '{"extends":"./missing.json"}']) {
		write(root, 'tsconfig.json', config);
		result = check(root);
		assert.equal(result.status, 1, config);
		assert.match(result.stderr, /tsconfig\.json.*configuration error/);
	}
	rmSync(path.join(root, 'tsconfig.json'));
	result = check(root);
	assert.equal(result.status, 1);
	assert.match(result.stderr, /tsconfig\.json.*configuration error/);
});

test('rejects new runtime cycles with paths and accepts their removal', t => {
	const root = createRepository(t);
	write(root, 'src/a.ts', "export * from '@/b';\n");
	write(root, 'src/b.ts', "import './a';\n");
	const result = check(root);
	assert.equal(result.status, 1);
	assert.match(result.stderr, /new runtime cycle: src\/a\.ts -> src\/b\.ts -> src\/a\.ts/);
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Cycle base');
	write(root, 'src/b.ts', 'export {};\n');
	assert.equal(check(root).status, 0);
});

test('reports inherited cycles but rejects them against an earlier explicit revision', t => {
	const root = createRepository(t);
	const earlier = git(root, 'rev-parse', 'HEAD');
	write(root, 'src/a.ts', "export * from '@/b';\n");
	write(root, 'src/b.ts', "import './a';\n");
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Inherited cycle');
	write(root, 'tsconfig.json', JSON.stringify({compilerOptions: {paths: {'alias/*': ['./src/*']}}}));
	write(root, 'src/a.ts', "export * from 'alias/b';\n");
	const inherited = check(root);
	assert.equal(inherited.status, 0, inherited.stderr);
	assert.match(inherited.stdout, /inherited runtime cycle: src\/a\.ts -> src\/b\.ts -> src\/a\.ts/);
	const introduced = check(root, '--base', earlier);
	assert.equal(introduced.status, 1);
	assert.match(introduced.stderr, /new runtime cycle/);
});

test('rejects new cycles within an inherited component and every inherited AI cycle', t => {
	const root = createRepository(t);
	write(root, 'src/a.ts', "import './b';\n");
	write(root, 'src/b.ts', "import './c';\n");
	write(root, 'src/c.ts', "import './a';\n");
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Inherited component');
	write(root, 'src/a.ts', "import './b'; import './c';\n");
	const added = check(root);
	assert.equal(added.status, 1);
	assert.match(added.stderr, /new runtime cycle: src\/a\.ts -> src\/c\.ts -> src\/a\.ts/);
	write(root, 'src/features/ai/a.ts', "import './b';\n");
	write(root, 'src/features/ai/b.ts', "import './a';\n");
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Inherited AI cycle');
	const ai = check(root);
	assert.equal(ai.status, 1);
	assert.match(ai.stderr, /AI runtime cycle: src\/features\/ai\/a\.ts -> src\/features\/ai\/b\.ts/);
});

test('excludes only type edges while preserving mixed values and side effects', t => {
	const root = createRepository(t);
	write(root, 'src/a.ts', "import type {B} from './b'; export type {B} from './b'; import {type B as Other} from './b'; export {type B as Again} from './b'; export type A = B;\n");
	write(root, 'src/b.ts', "import './a'; export type B = number; export const value = 1;\n");
	write(root, 'src/types.d.ts', "import './a';\n");
	const types = check(root);
	assert.equal(types.status, 0, types.stderr);
	assert.match(types.stdout, /1 runtime edges/);
	write(root, 'tsconfig.json', '{"compilerOptions":{"module":"esnext","moduleResolution":"bundler","verbatimModuleSyntax":true}}');
	for (const statement of ["import './b';", "import {type B, value} from './b'; console.log(value);", "export {type B, value} from './b';", "import {} from './b';"]) {
		write(root, 'src/a.ts', statement + '\n');
		const values = check(root);
		assert.equal(values.status, 1, statement);
		assert.match(values.stderr, /new runtime cycle/);
	}
});

test('enforces the AI public index for value and type imports while allowing internal and shared contracts', t => {
	const root = createRepository(t);
	write(root, 'src/features/ai/private.ts', "import type {Shared} from '@/shared/contracts'; export type Private = Shared; export const value = 1;\n");
	write(root, 'src/features/ai/index.ts', "export {value} from './private'; export type {Private} from './private';\n");
	write(root, 'src/shared/contracts.ts', 'export type Shared = number;\n');
	write(root, 'src/shell.ts', "import {value} from '@/features/ai'; import type {Private} from '@/features/ai';\n");
	assert.equal(check(root).status, 0);
	for (const statement of ["import {value} from '@/features/ai/private';", "import type {Private} from '@/features/ai/private';", "export type {Private} from '@/features/ai/private';"]) {
		write(root, 'src/shell.ts', statement + '\n');
		const result = check(root);
		assert.equal(result.status, 1, statement);
		assert.match(result.stderr, /src\/shell\.ts -> src\/features\/ai\/private\.ts.*public integration surface/);
	}
});

test('resolves existing local stylesheet and asset imports and rejects missing assets', t => {
	const root = createRepository(t);
	write(root, 'src/base.ts', "import './style.css'; import icon from '@/icon.svg'; export {icon};\n");
	write(root, 'src/style.css', 'body { color: red; }\n');
	write(root, 'src/icon.svg', '<svg/>\n');
	const valid = check(root);
	assert.equal(valid.status, 0, valid.stderr);
	assert.match(valid.stdout, /2 runtime edges/);
	rmSync(path.join(root, 'src/icon.svg'));
	const missing = check(root);
	assert.equal(missing.status, 1);
	assert.match(missing.stderr, /unresolved local import @\/icon\.svg/);
});

test('includes static CommonJS and import-equals edges without matching comments or strings', t => {
	const root = createRepository(t);
	write(root, 'tsconfig.json', '{"compilerOptions":{"module":"commonjs","moduleResolution":"node"}}');
	write(root, 'src/a.cts', "import b = require('./b.cjs'); export {b};\n");
	write(root, 'src/b.cjs', "function load() { return require('./a.cjs'); }\n");
	const runtime = check(root);
	assert.equal(runtime.status, 1);
	assert.match(runtime.stderr, /new runtime cycle: src\/a\.cts -> src\/b\.cjs -> src\/a\.cts/);
	write(root, 'src/a.cts', "// require('./b.cjs')\nconst source = `require('./b.cjs')`;\n");
	const inert = check(root);
	assert.equal(inert.status, 0, inert.stderr);
});

test('fails closed when required historical dependency inputs are unreadable or malformed', t => {
	const root = createRepository(t);
	write(root, 'tsconfig.json', '{"compilerOptions":{"moduleResolution":"invalid"}}');
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Broken base config');
	write(root, 'tsconfig.json', '{"compilerOptions":{}}');
	const invalid = check(root);
	assert.equal(invalid.status, 1);
	assert.match(invalid.stderr, /change base .*tsconfig\.json.*configuration error/);
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Fixed config');
	const blob = git(root, 'rev-parse', 'HEAD:src/base.ts');
	rmSync(path.join(root, '.git/objects', blob.slice(0, 2), blob.slice(2)));
	const unreadable = check(root);
	assert.equal(unreadable.status, 1);
	assert.match(unreadable.stderr, /change base .*Cannot read source .*src\/base\.ts/);
});
