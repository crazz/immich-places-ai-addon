import react from '@vitejs/plugin-react';
import {defineConfig} from 'vitest/config';

const aiSourceGlob = 'src/features/ai/**';

export default defineConfig({
	plugins: [react()],
	resolve: {tsconfigPaths: true},
	test: {
		environment: 'jsdom',
		include: ['src/**/*.test.{ts,tsx}'],
		setupFiles: ['tests/unit/setup.ts'],
		globals: false,
		allowOnly: false,
		passWithNoTests: false,
		retry: 0,
		coverage: {
			provider: 'v8',
			include: ['src/**/*.{ts,tsx}'],
			exclude: ['src/**/*.test.{ts,tsx}', 'src/**/*.d.ts'],
			reporter: ['text', 'html', 'json-summary'],
			thresholds: {
				[aiSourceGlob]: {lines: 80, branches: 80}
			}
		}
	}
});
