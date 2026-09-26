import assert from 'node:assert/strict';
import test from 'node:test';

import {writeFixtureEnvironment} from '../../tests/e2e/backend-control.mjs';

test('synthetic GPS dispatch requires an explicit disposable-fixture opt-in and resets disabled', () => {
 assert.deepEqual(writeFixtureEnvironment({}), {AI_WRITE_ENABLED: 'false', AI_WRITE_PROFILE: ''});
 assert.throws(() => writeFixtureEnvironment({writeEnabled: true}), /disposable/);
 assert.throws(() => writeFixtureEnvironment({writeEnabled: true, disposableFixture: 'private-photo'}), /disposable/);
 assert.deepEqual(writeFixtureEnvironment({writeEnabled: true, disposableFixture: 'synthetic-gps-only-v1'}), {AI_WRITE_ENABLED: 'true', AI_WRITE_PROFILE: 'immich-v3.2.2'});
 assert.deepEqual(writeFixtureEnvironment({writeEnabled: false}), {AI_WRITE_ENABLED: 'false', AI_WRITE_PROFILE: ''});
});

test('synthetic descriptions require separate installation-bound disposable authorization', () => {
 const installation = 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa';
 const input = {writeEnabled: true, descriptionEnabled: true, disposableFixture: 'synthetic-standard-fields-v2'};
 const environment = writeFixtureEnvironment(input, installation);
 assert.deepEqual(JSON.parse(environment.AI_WRITE_CAPABILITIES), {version: 1, installation, profile: 'immich-v3.2.2', evidence: 'loopback-standard-fields-v2', capabilities: ['description']});
 assert.throws(() => writeFixtureEnvironment({...input, writeEnabled: false}, installation), /disposable/);
 assert.throws(() => writeFixtureEnvironment(input), /installation/);
 assert.throws(() => writeFixtureEnvironment({...input, disposableFixture: 'synthetic-gps-only-v1'}, installation), /disposable/);
});

test('synthetic stack writes require separately named member-scope authorization', () => {
 const installation = 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa';
 const input = {writeEnabled: true, stackEnabled: true, disposableFixture: 'synthetic-stack-targets-v3'};
 assert.deepEqual(JSON.parse(writeFixtureEnvironment(input, installation).AI_WRITE_CAPABILITIES), {version: 1, installation, profile: 'immich-v3.2.2', evidence: 'loopback-stack-targets-v3', capabilities: ['stack_gps']});
 assert.deepEqual(JSON.parse(writeFixtureEnvironment({...input, descriptionEnabled: true}, installation).AI_WRITE_CAPABILITIES).capabilities, ['description', 'stack_gps']);
 assert.throws(() => writeFixtureEnvironment(input), /installation/);
 assert.throws(() => writeFixtureEnvironment({...input, writeEnabled: false}, installation), /disposable/);
 assert.throws(() => writeFixtureEnvironment({...input, disposableFixture: 'synthetic-standard-fields-v2'}, installation), /disposable/);
 assert.throws(() => writeFixtureEnvironment({...input, disposableFixture: 'synthetic-gps-only-v1'}, installation), /disposable/);
});
