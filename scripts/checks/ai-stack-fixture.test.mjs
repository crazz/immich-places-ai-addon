import assert from 'node:assert/strict';
import test from 'node:test';

import {makeStackFixture} from '../../tests/e2e/stack-fixture.mjs';

test('synthetic stack scope has five explicit eligible members and preserves unrelated assets', () => {
 const primary = '11111111-1111-4111-8111-111111111111';
 const assets = [{id: primary, type: 'IMAGE', exifInfo: {latitude: null, longitude: null}}, {id: 'unrelated', type: 'IMAGE', exifInfo: {latitude: null, longitude: null}}];
 const before = structuredClone(assets);
 const fixture = makeStackFixture(assets);
 assert.deepEqual(assets, before);
 assert.equal(fixture.stack.primaryAssetId, primary);
 assert.equal(fixture.stack.assets.length, 5);
 assert.equal(new Set(fixture.stack.assets.map(asset => asset.id)).size, 5);
 assert.ok(fixture.stack.assets.every(asset => asset.type === 'IMAGE' && asset.visibility === 'timeline' && asset.isTrashed === false && asset.stack.id === fixture.stack.id && asset.stack.assetCount === 5));
 assert.equal(fixture.assets.length, 6);
 assert.deepEqual(fixture.assets.find(asset => asset.id === 'unrelated'), before[1]);
 assert.ok(fixture.stack.assets.slice(1).every(asset => asset.exifInfo.description === 'Preserved sibling description'));
 assert.throws(() => makeStackFixture([]), /primary/);
});
