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
