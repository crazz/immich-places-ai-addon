import assert from 'node:assert/strict';
import {spawnSync} from 'node:child_process';
import {mkdirSync, writeFileSync} from 'node:fs';
import path from 'node:path';

export function fixtureEnvironment() {
	const environment = Object.fromEntries(Object.entries(process.env).filter(([name]) => !name.startsWith('GIT_')));
	return {...environment, GIT_CONFIG_GLOBAL: '/dev/null', GIT_CONFIG_NOSYSTEM: '1', GIT_TERMINAL_PROMPT: '0'};
}

export function git(root, ...args) {
	const result = spawnSync('git', args, {cwd: root, encoding: 'utf8', env: fixtureEnvironment()});
	assert.equal(result.status, 0, result.stderr);
	return result.stdout.trim();
}

export function write(root, name, content) {
	mkdirSync(path.dirname(path.join(root, name)), {recursive: true});
	writeFileSync(path.join(root, name), content);
}
