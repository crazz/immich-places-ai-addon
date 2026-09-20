import {spawn} from 'node:child_process';
import {once} from 'node:events';
import {mkdtempSync, rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import path from 'node:path';
import {setTimeout as delay} from 'node:timers/promises';
import {fileURLToPath} from 'node:url';

import {startBackendControl} from './backend-control.mjs';

const dataDir = mkdtempSync(path.join(tmpdir(), 'immich-places-smoke-'));
const binary = fileURLToPath(new URL('../../out/checks/immich-places-backend', import.meta.url));
const aiEnabled = process.env.SMOKE_AI_ENABLED === 'true';
const options = {
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
		AI_ENABLED: aiEnabled ? 'true' : 'false',
		AI_PUBLIC_ORIGIN: 'http://127.0.0.1:3080',
		AI_PROVIDER_EGRESS_POLICY: aiEnabled
			? '[{"baseURL":"http://127.0.0.1:8090/v1","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]'
			: '[]',
		SYNC_INTERVAL_MS: '3600000',
		HTTP_PROXY: 'http://127.0.0.1:8090',
		HTTPS_PROXY: 'http://127.0.0.1:8090',
		NO_PROXY: '127.0.0.1,localhost'
	}
};

let isRestarting = false;
let control;
let child;
function start() {
 const processChild = spawn(binary, [], options);
 processChild.on('error', error => {console.error(error); process.exitCode = 1; control?.close();});
 processChild.on('exit', code => {
  if (!isRestarting) {process.exitCode = code ?? 1; control?.close();}
 });
 return processChild;
}
child = start();
if (aiEnabled) {
 control = startBackendControl(dataDir, async update => {
  isRestarting = true;
  try {
   const exited = once(child, 'exit');
   child.kill('SIGTERM');
   await exited;
   Object.assign(options.env, update);
   child = start();
   const deadline = Date.now() + 10000;
   while (Date.now() < deadline) {
    const response = await fetch('http://127.0.0.1:8089/health', {signal: AbortSignal.timeout(500)}).catch(() => null);
    if (response?.ok) {return;}
    await delay(50);
   }
   throw new Error('Synthetic backend restart timed out');
  } finally {isRestarting = false;}
 });
}

process.on('SIGTERM', () => {control?.close(); child.kill('SIGTERM');});
process.on('SIGINT', () => {control?.close(); child.kill('SIGINT');});
process.on('exit', () => rmSync(dataDir, {recursive: true, force: true}));
