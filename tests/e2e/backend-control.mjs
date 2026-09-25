import {createServer} from 'node:http';
import path from 'node:path';
import {DatabaseSync} from 'node:sqlite';

export function writeFixtureEnvironment(input) {
 if (input.writeEnabled === true && input.disposableFixture !== 'synthetic-gps-only-v1') {throw new Error('Explicit disposable fixture authorization required');}
 return {AI_WRITE_ENABLED: input.writeEnabled === true ? 'true' : 'false', AI_WRITE_PROFILE: input.writeEnabled === true ? 'immich-v3.2.2' : ''};
}

export function startBackendControl(dataDir, restart) {
 const policies = new Map();
 let isBusy = false;
 const server = createServer(async (request, response) => {
  if (request.method !== 'POST' || request.url !== '/configure' || isBusy) {
   response.writeHead(409); response.end(); return;
  }
  isBusy = true;
  try {
   let raw = '';
   for await (const chunk of request) {
    raw += chunk;
    if (raw.length > 2048) {throw new Error('Oversized fixture configuration');}
   }
   const input = JSON.parse(raw);
   if (input.profile) {
    const db = new DatabaseSync(path.join(dataDir, 'immich-places.db'), {readOnly: true});
    try {
     const row = db.prepare(`SELECT p.userID AS owner, p.id AS profile, p.activeRevision AS revision,
      v.model, c.policyFingerprint AS egressFingerprint, i.id AS installation
      FROM ai_provider_profiles p JOIN ai_provider_versions v ON v.userID=p.userID AND v.profileID=p.id AND v.revision=p.activeRevision
      JOIN ai_provider_capability_checks c ON c.userID=p.userID AND c.profileID=p.id AND c.revision=p.activeRevision
      CROSS JOIN ai_installation_identity i WHERE p.userID=? AND p.id=? AND c.lifecycle='completed' AND i.singleton=1`).get(input.owner, input.profile);
     if (!row) {throw new Error('No completed synthetic provider binding');}
     policies.set(`${row.owner}:${row.profile}`, {
      version: 'execution-v1',
binding: row,
outputField: 'max_tokens',
context: true,
      maxInputTokens: 100000,
maxOutputTokens: 4000,
maxRequestBytes: 300000,
maxImageBytes: 65536,
      evidenceRef: 'loopback-browser-fixture-v1'
     });
    } finally {db.close();}
   }
   await restart({AI_EXECUTION_POLICIES: JSON.stringify([...policies.values()]), AI_ENABLED: input.enabled === false ? 'false' : 'true', ...writeFixtureEnvironment(input)});
   response.writeHead(200, {'Content-Type': 'application/json'}); response.end('{"ready":true}');
  } catch (error) {
   response.writeHead(500, {'Content-Type': 'application/json'}); response.end(JSON.stringify({error: error.message}));
  } finally {isBusy = false;}
 });
 server.requestTimeout = 15000;
 server.headersTimeout = 5000;
 server.listen(8091, '127.0.0.1');
 return server;
}
