import assert from 'node:assert/strict';
import {mkdtempSync, readFileSync, rmSync, writeFileSync} from 'node:fs';
import {tmpdir} from 'node:os';
import path from 'node:path';
import test from 'node:test';

test('enforces AI statement coverage across core, adapters and untested files', async () => {
	const {aiCoverage} = await import('./go-coverage.mjs').catch(() => ({}));
	assert.equal(typeof aiCoverage, 'function', 'AI coverage enforcement must exist');
	const measured = aiCoverage(`mode: atomic
immich-places-backend/legacy.go:1.1,9.1 100 0
immich-places-backend/internal/ai/providers/config.go:1.1,9.1 40 1
immich-places-backend/aiProviderHTTP.go:1.1,9.1 40 2
immich-places-backend/internal/aiadapters/unused.go:1.1,9.1 20 0
`);
	assert.deepEqual(measured, {covered: 80, total: 100, percent: 80});
	assert.throws(() => aiCoverage('mode: atomic\nimmich-places-backend/aiProviderHTTP.go:1.1,9.1 10 0\n'), /below 80%/);
	assert.throws(() => aiCoverage('mode: atomic\nimmich-places-backend/legacy.go:1.1,9.1 10 10\n'), /No AI statements/);
	assert.throws(() => aiCoverage('mode: atomic\ninvalid'), /Malformed coverage/);
});

test('runs one race suite with all-package instrumentation and fails closed on execution errors', async t => {
	const {runGoTests} = await import('./go-tests.mjs').catch(() => ({}));
	assert.equal(typeof runGoTests, 'function', 'race and coverage gate must exist');
	const root = mkdtempSync(path.join(tmpdir(), 'go-coverage-check-'));
	t.after(() => rmSync(root, {recursive: true, force: true}));
	const calls = [];
	const report = [];
	const execute = (command, args, options) => {
		calls.push({command, args, options});
		const profile = args.find(arg => arg.startsWith('-coverprofile=')).split('=')[1];
		writeFileSync(profile, 'mode: atomic\nimmich-places-backend/aiProviderHTTP.go:1.1,9.1 10 1\n');
		return {status: 0};
	};
	assert.equal(runGoTests({root, execute, report: line => report.push(line)}), 0);
	assert.equal(calls.length, 1);
	assert.equal(calls[0].command, 'go');
	assert.deepEqual(calls[0].args.filter(arg => !arg.startsWith('-coverprofile=')), ['test', '-mod=readonly', '-race', '-coverpkg=./...', './...']);
	assert.equal(calls[0].options.cwd, path.join(root, 'backend'));
	assert.equal(calls[0].options.shell, false);
	assert.match(report.join('\n'), /AI Go statement coverage: 100.00%/);
	assert.match(readFileSync(path.join(root, 'out/checks/backend-coverage.out'), 'utf8'), /mode: atomic/);
	for (const outcome of [{status: 1}, {status: null, error: new Error('missing go')}, {status: 0, signal: 'SIGTERM'}]) {
		assert.throws(() => runGoTests({root, execute: () => outcome}), /Go tests failed/);
	}
	assert.throws(() => runGoTests({root, execute: () => ({status: 0})}), /ENOENT/);
});
