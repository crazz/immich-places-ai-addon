import {afterEach, expect, it, vi} from 'vitest';

import {savedPreview, stagedPreviewDraft} from './testing/writePreview';
import {createWritePreview, fetchWritePreview} from './writePreviewApi';

afterEach(() => vi.unstubAllGlobals());

it('rejects malformed or substituted comparisons at the private response boundary', async () => {
 const draft = stagedPreviewDraft();
 const original = savedPreview();
 const foreign = '66666666-6666-4666-8666-666666666666';
 for (const plan of [
  {...original.plan, owner: 'another-owner'},
  {...original.plan, draftId: foreign},
  {...original.plan, draftRevision: 9},
  {...original.plan, analysisId: foreign},
  {...original.plan, targetId: foreign},
  {...original.plan, intended: {latitude: 1, longitude: 12}},
  {...original.plan, before: {latitude: null, longitude: null}},
  {...original.plan, imageIdentity: `v1:${'c'.repeat(64)}`},
  {...original.plan, fields: ['gps', 'description']},
  {...original.plan, before: {latitude: 91, longitude: null}},
  {...original.plan, expiresAt: original.plan.createdAt}
 ]) {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({...original, plan}))));
  await expect(createWritePreview('owner', draft)).rejects.toThrow();
 }
 vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify(original))));
 await expect(fetchWritePreview('owner', draft, foreign)).rejects.toThrow('Preview reference changed');
});
