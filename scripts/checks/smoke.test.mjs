import assert from 'node:assert/strict';
import test from 'node:test';

test('runs disabled legacy and enabled provider journeys and propagates either failure', async () => {
	const {runSmoke} = await import('./smoke.mjs').catch(() => ({}));
	assert.equal(typeof runSmoke, 'function', 'smoke gate must verify both installation modes');
	for (const failed of [undefined, 'false', 'true']) {
		const calls = [];
		const status = runSmoke((command, args, options) => {
			calls.push({command, args, options});
			return {status: options.env.SMOKE_AI_ENABLED === failed ? 1 : 0};
		});
		assert.equal(status, failed ? 1 : 0);
		assert.deepEqual(calls.map(call => call.options.env.SMOKE_AI_ENABLED), ['false', 'true']);
		assert.ok(calls.every(call => call.command === process.execPath && call.options.shell === false));
		assert.deepEqual(calls[0].args, ['node_modules/@playwright/test/cli.js', 'test']);
	}
	assert.equal(runSmoke(() => ({status: null, signal: 'SIGTERM'})), 1);
});
