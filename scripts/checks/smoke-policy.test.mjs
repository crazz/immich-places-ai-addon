import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import test from 'node:test';
import {fileURLToPath} from 'node:url';

import {isHandwrittenSource} from './source-scope.mjs';
import {verify} from './verification-runner.mjs';

test('counts executable browser fixtures while excluding Playwright-generated reports and results', () => {
	assert.equal(isHandwrittenSource('tests/e2e/immich-server.mjs', 'const fixture = true;'), true);
	assert.equal(isHandwrittenSource('playwright-report/assets/report.js', 'generated();'), false);
	assert.equal(isHandwrittenSource('test-results/trace/resources/page.js', 'generated();'), false);
});

test('AI-enabled smoke backend configures an explicit loopback provider policy', async () => {
	const source = await readFile(new URL('../../tests/e2e/backend-server.mjs', import.meta.url), 'utf8');
	assert.match(source, /AI_PROVIDER_EGRESS_POLICY/);
	assert.match(source, /127\.0\.0\.1:8090\/v1/);
	assert.match(source, /127\.0\.0\.0\/8/);
	assert.match(source, /aiEnabled \? 'true' : 'false'/);
});

test('a browser failure makes focused verification fail after both builds pass', () => {
	const lines = [];
	const status = verify({
		root: fileURLToPath(new URL('../../', import.meta.url)),
		base: 'HEAD',
		gate: 'smoke',
		execute: (command, args) => ({status: args[0] === 'scripts/checks/smoke.mjs' ? 3 : 0}),
		report: line => lines.push(line)
	});
	assert.equal(status, 1);
	assert.deepEqual(lines, ['RUN go-build', 'PASS go-build', 'RUN frontend-build', 'PASS frontend-build', 'RUN smoke', 'FAIL smoke: exit 3']);
});
