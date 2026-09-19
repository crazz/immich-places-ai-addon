import {spawn} from 'node:child_process';
import {mkdtempSync, rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import path from 'node:path';
import {fileURLToPath} from 'node:url';

const dataDir = mkdtempSync(path.join(tmpdir(), 'immich-places-smoke-'));
const binary = fileURLToPath(new URL('../../out/checks/immich-places-backend', import.meta.url));
const child = spawn(binary, [], {
	cwd: dataDir,
	stdio: 'inherit',
	// Do not inherit a developer's service configuration or load their dotenv files.
	env: {
		PATH: process.env.PATH,
		TZ: 'UTC',
		PORT: '8089',
		DATA_DIR: dataDir,
		IMMICH_URL: 'http://127.0.0.1:8090',
		ENCRYPTION_KEY: 'synthetic-smoke-encryption-key',
		ALLOW_INSECURE: 'true',
		REGISTRATION_ENABLED: 'true',
		AI_ENABLED: process.env.SMOKE_AI_ENABLED === 'true' ? 'true' : 'false',
		AI_PUBLIC_ORIGIN: 'http://127.0.0.1:3080',
		SYNC_INTERVAL_MS: '3600000',
		HTTP_PROXY: 'http://127.0.0.1:8090',
		HTTPS_PROXY: 'http://127.0.0.1:8090',
		NO_PROXY: '127.0.0.1,localhost'
	}
});

process.on('SIGTERM', () => child.kill('SIGTERM'));
process.on('SIGINT', () => child.kill('SIGINT'));
process.on('exit', () => rmSync(dataDir, {recursive: true, force: true}));
child.on('error', error => {
	console.error(error);
	process.exitCode = 1;
});
child.on('exit', code => {
	process.exitCode = code ?? 1;
});
