import assert from 'node:assert/strict';
import test from 'node:test';

import {writeFixtureEnvironment} from '../../tests/e2e/backend-control.mjs';

test('synthetic metadata requires its own explicit installation-bound disposable authorization', () => {
 const installation = 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa';
 const input = {writeEnabled: true, metadataEnabled: true, disposableFixture: 'synthetic-metadata-v4'};
 assert.deepEqual(JSON.parse(writeFixtureEnvironment(input, installation).AI_WRITE_CAPABILITIES), {version: 1, installation, profile: 'immich-v3.2.2', evidence: 'loopback-metadata-v4', capabilities: ['metadata']});
 assert.deepEqual(JSON.parse(writeFixtureEnvironment({...input, descriptionEnabled: true, stackEnabled: true}, installation).AI_WRITE_CAPABILITIES).capabilities, ['description', 'stack_gps', 'metadata']);
 for (const fixture of ['synthetic-gps-only-v1', 'synthetic-standard-fields-v2', 'synthetic-stack-targets-v3']) {
  assert.throws(() => writeFixtureEnvironment({...input, disposableFixture: fixture}, installation), /disposable/);
 }
 assert.throws(() => writeFixtureEnvironment(input), /installation/);
 assert.throws(() => writeFixtureEnvironment({...input, writeEnabled: false}, installation), /disposable/);
 assert.deepEqual(writeFixtureEnvironment({}), {AI_WRITE_ENABLED: 'false', AI_WRITE_PROFILE: ''});
});
