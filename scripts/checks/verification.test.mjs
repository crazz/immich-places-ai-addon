import assert from 'node:assert/strict';
import {spawnSync} from 'node:child_process';
import {existsSync, mkdirSync, mkdtempSync, rmSync, symlinkSync} from 'node:fs';
import {tmpdir} from 'node:os';
import path from 'node:path';
import test from 'node:test';
import {fileURLToPath} from 'node:url';

const cli = fileURLToPath(new URL('./verify.mjs', import.meta.url));

test('provides verification usage and rejects malformed arguments', () => {
	const help = spawnSync(process.execPath, [cli, '--help'], {encoding: 'utf8'});
	assert.equal(help.status, 0, 'verification help must succeed');
	assert.match(help.stdout, /Usage:.*--base <revision>.*--gate <name>/);
	assert.match(help.stdout, /checker-tests.*size.*dependencies.*gofmt.*lint.*typegen.*types.*go-vet.*go-tests.*go-build.*frontend-build/s);
	for (const args of [['--base'], ['--gate'], ['--unknown'], ['--gate', 'unknown'], ['--base', 'HEAD', '--base', 'HEAD~1']]) {
		const result = spawnSync(process.execPath, [cli, ...args], {encoding: 'utf8'});
		assert.equal(result.status, 1);
		assert.match(result.stderr, /Usage:|Unknown gate/);
	}
});

test('runs every required gate with safe arguments and visible passing statuses', async () => {
	const {verify} = await import('./verification-runner.mjs').catch(() => ({}));
	assert.equal(typeof verify, 'function', 'shared verification runner must exist');
	const calls = [];
	const lines = [];
	const root = fileURLToPath(new URL('../../', import.meta.url));
	const result = verify({
root,
base: 'base;literal',
execute: (command, args, options) => {
		calls.push({command, args, options});
		return {status: 0};
	},
report: line => lines.push(line)
});
	assert.equal(result, 0);
	const names = ['checker-tests', 'size', 'dependencies', 'gofmt', 'lint', 'typegen', 'types', 'go-vet', 'go-tests', 'go-build', 'frontend-tests', 'frontend-build', 'smoke'];
	assert.deepEqual(lines.filter(line => line.startsWith('PASS ')), names.map(name => `PASS ${name}`));
	assert.equal(calls.length, names.length);
	assert.deepEqual(calls[1].args, ['scripts/checks/size.mjs', '--base', 'base;literal']);
	assert.deepEqual(calls[2].args, ['scripts/checks/dependencies.mjs', '--base', 'base;literal']);
	assert.deepEqual(calls[5].args, ['node_modules/next/dist/bin/next', 'typegen']);
	assert.deepEqual(calls[6].args, ['node_modules/typescript/bin/tsc', '--noEmit']);
	assert.deepEqual(calls[8].args, ['test', '-mod=readonly', '-race', './...']);
	assert.equal(calls[8].options.cwd, `${root}backend`);
	assert.equal(calls[9].args.at(-1), './...');
	assert.match(calls[9].args.at(-2), /out\/checks\/immich-places-backend$/);
	assert.equal(calls[10].command, process.execPath);
	assert.deepEqual(calls[10].args, ['node_modules/vitest/vitest.mjs', 'run', '--coverage']);
	assert.equal(calls[10].options.cwd, root);
	assert.deepEqual(calls[12].args, ['node_modules/@playwright/test/cli.js', 'test']);
	assert.ok(calls.every(call => call.options.shell === false && call.options.stdio === 'inherit'));
	assert.ok(calls[0].args.includes('scripts/checks/verification.test.mjs'));
});

test('reports nonzero exits and continues independent gates without an overall pass', async () => {
	const {verify} = await import('./verification-runner.mjs');
	const lines = [];
	const result = verify({root: fileURLToPath(new URL('../../', import.meta.url)), base: 'HEAD', execute: (command, args) => ({status: args[0].includes('eslint') ? 7 : 0}), report: line => lines.push(line)});
	assert.equal(result, 1);
	assert.ok(lines.includes('FAIL lint: exit 7'), lines.join('\n'));
	assert.ok(lines.includes('PASS frontend-build'));
	assert.ok(!lines.includes('PASS lint'));
});

test('reports missing commands and execution exceptions as named failures', async () => {
	const {verify} = await import('./verification-runner.mjs');
	for (const throws of [false, true]) {
		const lines = [];
		const result = verify({
root: fileURLToPath(new URL('../../', import.meta.url)),
base: 'HEAD',
execute: command => {
			if (command === 'go') {
				const error = new Error('spawn go ENOENT');
				if (throws) {
					throw error;
				}
				return {status: null, error};
			}
			return {status: 0};
		},
report: line => lines.push(line)
});
		assert.equal(result, 1);
		assert.ok(lines.includes('FAIL go-vet: spawn go ENOENT'), lines.join('\n'));
		assert.ok(lines.includes('FAIL go-tests: spawn go ENOENT'));
		assert.ok(lines.includes('PASS frontend-build'));
	}
});

test('reports termination signals and never treats absent exit status as success', async () => {
	const {verify} = await import('./verification-runner.mjs');
	for (const outcome of [{status: null, signal: 'SIGTERM'}, {status: 0, signal: 'SIGKILL'}, {status: null}]) {
		const lines = [];
		const result = verify({root: fileURLToPath(new URL('../../', import.meta.url)), base: 'HEAD', execute: (command, args) => args[0].includes('eslint') ? outcome : {status: 0}, report: line => lines.push(line)});
		assert.equal(result, 1);
		assert.ok(lines.includes(`FAIL lint: ${outcome.signal ? `terminated by ${outcome.signal}` : 'exit null'}`), lines.join('\n'));
		assert.ok(!lines.includes('PASS lint'));
	}
});

test('focuses one gate while generating route types before a standalone typecheck', async () => {
	const {verify} = await import('./verification-runner.mjs');
	for (const [gate, expected] of [['lint', ['lint']], ['types', ['typegen', 'types']]]) {
		const lines = [];
		assert.equal(verify({root: fileURLToPath(new URL('../../', import.meta.url)), base: 'HEAD', gate, execute: () => ({status: 0}), report: line => lines.push(line)}), 0);
		assert.deepEqual(lines.filter(line => line.startsWith('PASS ')), expected.map(name => `PASS ${name}`));
	}
});

test('blocks typechecking when route generation fails and reports independent gates', async () => {
	const {verify} = await import('./verification-runner.mjs');
	const lines = [];
	const calls = [];
	assert.equal(verify({
root: fileURLToPath(new URL('../../', import.meta.url)),
base: 'HEAD',
execute: (command, args) => {
		calls.push(args);
		return {status: args[1] === 'typegen' ? 2 : 0};
	},
report: line => lines.push(line)
}), 1);
	assert.ok(lines.includes('BLOCKED types: typegen failed'), lines.join('\n'));
	assert.ok(!calls.some(args => args[0].includes('typescript')));
	assert.ok(lines.includes('PASS frontend-build'));
});

test('resolves an explicit base and propagates real command failures through the CLI', t => {
	const invalid = spawnSync(process.execPath, [cli, '--base', 'missing;revision', '--gate', 'lint'], {encoding: 'utf8'});
	assert.equal(invalid.status, 1);
	assert.match(invalid.stderr, /Invalid change base: missing;revision/);
	const tools = mkdtempSync(path.join(tmpdir(), 'verification-missing-go-'));
	t.after(() => rmSync(tools, {recursive: true, force: true}));
	const git = process.env.PATH.split(path.delimiter).map(directory => path.resolve(directory, 'git')).find(existsSync);
	assert.ok(git, 'Git must be available for the comparison-base check');
	symlinkSync(git, path.join(tools, 'git'));
	const failed = spawnSync(process.execPath, [cli, '--base', 'HEAD', '--gate', 'go-vet'], {encoding: 'utf8', env: {...process.env, PATH: tools}});
	assert.equal(failed.status, 1, 'missing Go executable must fail the real CLI');
	assert.match(failed.stdout, /base: [a-f0-9]{40}/);
	assert.match(failed.stdout, /FAIL go-vet:.*ENOENT/);
});

test('rejects an empty checker test inventory instead of allowing an empty pass', async t => {
	const {verify} = await import('./verification-runner.mjs');
	const root = mkdtempSync(path.join(tmpdir(), 'verification-empty-'));
	t.after(() => rmSync(root, {recursive: true, force: true}));
	mkdirSync(path.join(root, 'scripts/checks'), {recursive: true});
	assert.throws(() => verify({root, base: 'HEAD', execute: () => ({status: 0}), report: () => {}}), /No checker tests found/);
});

test('propagates frontend test failures through a focused verification run', async () => {
	const {verify} = await import('./verification-runner.mjs');
	const lines = [];
	const calls = [];
	const result = verify({
		root: fileURLToPath(new URL('../../', import.meta.url)),
		base: 'HEAD',
		gate: 'frontend-tests',
		execute: (command, args) => {
			calls.push({command, args});
			return {status: 1};
		},
		report: line => lines.push(line)
	});
	assert.equal(result, 1);
	assert.deepEqual(calls, [{command: process.execPath, args: ['node_modules/vitest/vitest.mjs', 'run', '--coverage']}]);
	assert.deepEqual(lines, ['RUN frontend-tests', 'FAIL frontend-tests: exit 1']);
});

test('focused smoke execution builds both artifacts before running the browser', async () => {
	const {verify} = await import('./verification-runner.mjs');
	const lines = [];
	const result = verify({
		root: fileURLToPath(new URL('../../', import.meta.url)),
		base: 'HEAD',
		gate: 'smoke',
		execute: () => ({status: 0}),
		report: line => lines.push(line)
	});
	assert.equal(result, 0);
	assert.deepEqual(lines, ['RUN go-build', 'PASS go-build', 'RUN frontend-build', 'PASS frontend-build', 'RUN smoke', 'PASS smoke']);
});

test('blocks smoke after either build fails while retaining independent results', async () => {
	const {verify} = await import('./verification-runner.mjs');
	for (const broken of ['go-build', 'frontend-build']) {
		for (const gate of [undefined, 'smoke']) {
			const lines = [];
			const calls = [];
			const result = verify({
				root: fileURLToPath(new URL('../../', import.meta.url)),
				base: 'HEAD',
				gate,
				execute: (command, args) => {
					calls.push(args);
					const fails = broken === 'go-build' ? command === 'go' && args[0] === 'build' : args[1] === 'build';
					return {status: fails ? 1 : 0};
				},
				report: line => lines.push(line)
			});
			assert.equal(result, 1);
			assert.ok(lines.includes(`BLOCKED smoke: ${broken} failed`), lines.join('\n'));
			assert.ok(!calls.some(args => args[0].includes('@playwright')));
			assert.ok(lines.includes(`PASS ${broken === 'go-build' ? 'frontend-build' : 'go-build'}`));
		}
	}
});
