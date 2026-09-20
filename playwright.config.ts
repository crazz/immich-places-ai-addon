import {defineConfig} from '@playwright/test';

const isAIEnabled = process.env.SMOKE_AI_ENABLED === 'true';
const mode = isAIEnabled ? 'ai-enabled' : 'ai-disabled';

const serverDefaults = {
	reuseExistingServer: false,
	timeout: 30_000,
	gracefulShutdown: {signal: 'SIGTERM' as const, timeout: 8_000}
};

export default defineConfig({
	testDir: './tests/e2e',
	testMatch: isAIEnabled ? ['providers.spec.ts', 'workflows.spec.ts', 'history.spec.ts', 'review.spec.ts'] : ['auth.spec.ts', 'manual.spec.ts', 'gpx.spec.ts'],
	outputDir: `test-results/${mode}`,
	forbidOnly: true,
	fullyParallel: false,
	workers: 1,
	retries: 0,
	timeout: 30_000,
	globalTimeout: 180_000,
	expect: {timeout: 10_000},
	reporter: [['list'], ['html', {open: 'never', outputFolder: `playwright-report/${mode}`}]],
	use: {
		baseURL: 'http://127.0.0.1:3080',
		browserName: 'chromium',
		viewport: {width: 1440, height: 1000},
		locale: 'en-US',
		timezoneId: 'UTC',
		serviceWorkers: 'block',
		trace: 'retain-on-failure',
		screenshot: 'only-on-failure'
	},
	webServer: [
		{...serverDefaults, name: 'Fake Immich', command: 'node tests/e2e/immich-server.mjs', url: 'http://127.0.0.1:8090/health'},
		{...serverDefaults, name: 'Go backend', command: 'node tests/e2e/backend-server.mjs', url: 'http://127.0.0.1:8089/health'},
		{
			...serverDefaults,
			name: 'Next.js',
			command: 'node tests/e2e/frontend-server.mjs',
			url: 'http://127.0.0.1:3080',
			env: Object.fromEntries([
				['BACKEND_URL', 'http://127.0.0.1:8089'], ['NEXT_TELEMETRY_DISABLED', '1'],
				['HOSTNAME', '127.0.0.1'], ['PORT', '3080']
			])
		}
	]
});
