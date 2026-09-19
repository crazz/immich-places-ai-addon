import {spawnSync} from 'node:child_process';
import {copyFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import path from 'node:path';
import {fileURLToPath} from 'node:url';

import {aiCoverage} from './go-coverage.mjs';

export function runGoTests({root, execute = spawnSync, report = console.log}) {
	const temporary = mkdtempSync(path.join(tmpdir(), 'immich-go-coverage-'));
	const profile = path.join(temporary, 'coverage.out');
	try {
		const result = execute('go', ['test', '-mod=readonly', '-race', '-coverpkg=./...', `-coverprofile=${profile}`, './...'], {
			cwd: path.join(root, 'backend'), stdio: 'inherit', shell: false
		});
		if (result.error || result.signal || result.status !== 0) {
			throw new Error(`Go tests failed: ${result.error?.message ?? result.signal ?? `exit ${result.status}`}`);
		}
		const output = path.join(root, 'out/checks');
		mkdirSync(output, {recursive: true});
		copyFileSync(profile, path.join(output, 'backend-coverage.out'));
		const coverage = aiCoverage(readFileSync(profile, 'utf8'));
		report(`AI Go statement coverage: ${coverage.percent.toFixed(2)}% (${coverage.covered}/${coverage.total}); minimum 80%`);
		return 0;
	} finally {
		rmSync(temporary, {recursive: true, force: true});
	}
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
	try {
		process.exitCode = runGoTests({root: process.cwd()});
	} catch (error) {
		console.error(error.message);
		process.exitCode = 1;
	}
}
