import {spawnSync} from 'node:child_process';
import path from 'node:path';
import {fileURLToPath} from 'node:url';

export function runSmoke(execute = spawnSync) {
	let status = 0;
	for (const isEnabled of ['false', 'true']) {
		const result = execute(process.execPath, ['node_modules/@playwright/test/cli.js', 'test'], {
			stdio: 'inherit', shell: false, env: {...process.env, SMOKE_AI_ENABLED: isEnabled}
		});
		if (result.status !== 0 || result.error || result.signal) {
			status = 1;
		}
	}
	return status;
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
	process.exitCode = runSmoke();
}
