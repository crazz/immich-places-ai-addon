import assert from 'node:assert/strict';
import {spawnSync} from 'node:child_process';
import {mkdtempSync, readFileSync, rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import path from 'node:path';
import test from 'node:test';
import {fileURLToPath} from 'node:url';

import {fixtureEnvironment, git, write} from './git-fixture.mjs';

const cli = fileURLToPath(new URL('./format-go.mjs', import.meta.url));

test('reports gofmt differences in tracked and new Go files without rewriting them', t => {
	const root = mkdtempSync(path.join(tmpdir(), 'format-go-'));
	t.after(() => rmSync(root, {recursive: true, force: true}));
	git(root, 'init', '-q');
	const content = 'package main\nfunc main( ) {}\n';
	write(root, 'backend/tracked.go', content);
	git(root, 'add', '.');
	write(root, 'backend/new file.go', content);
	write(root, '.gitignore', 'ignored.go\n');
	write(root, 'ignored.go', content);
	const result = spawnSync(process.execPath, [cli], {cwd: root, encoding: 'utf8', env: fixtureEnvironment()});
	assert.equal(result.status, 1);
	assert.match(result.stderr, /gofmt required: backend\/tracked.go/);
	assert.match(result.stderr, /gofmt required: backend\/new file.go/);
	assert.ok(!result.stderr.includes('ignored.go'));
	assert.equal(readFileSync(path.join(root, 'backend/tracked.go'), 'utf8'), content);
	assert.equal(readFileSync(path.join(root, 'backend/new file.go'), 'utf8'), content);
});

test('provides focused formatting help and rejects unexpected arguments', () => {
	const help = spawnSync(process.execPath, [cli, '--help'], {encoding: 'utf8'});
	assert.equal(help.status, 0);
	assert.match(help.stdout, /Usage: node scripts\/checks\/format-go.mjs/);
	const unknown = spawnSync(process.execPath, [cli, '--write'], {encoding: 'utf8'});
	assert.equal(unknown.status, 1);
	assert.match(unknown.stderr, /Usage:/);
});
