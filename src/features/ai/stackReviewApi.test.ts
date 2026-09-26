import {afterEach, expect, it, vi} from 'vitest';

import {createStackReview} from './stackReviewApi';
import {stackIDs, stackReview,stagedPreviewDraft} from './testing/writePreview';

afterEach(() => vi.unstubAllGlobals());

it('rejects incomplete foreign duplicate and stale target observations at the private review boundary', async () => {
 const draft = stagedPreviewDraft(); const ids = [draft.assetId, stackIDs[0]]; const review = stackReview(ids);
 for (const invalid of [
  {...review, owner: 'foreign'}, {...review, draftRevision: 9}, {...review, targets: review.targets.slice(1)},
  {...review, targets: [review.targets[0], review.targets[0]]},
  {...review, targets: [...review.targets].reverse()},
  {...review, candidates: [...review.candidates, review.candidates[0]]},
  {...review, targets: review.targets.map((target, index) => index === 0 ? {...target, imageIdentity: `v1:${'f'.repeat(64)}`} : target)},
  {...review, expiresAt: new Date(Date.now() - 1).toISOString()}
 ]) {
  vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(invalid))));
  await expect(createStackReview('owner', draft, ids)).rejects.toThrow();
 }
 const request = vi.fn(async () => new Response(JSON.stringify(review))); vi.stubGlobal('fetch', request);
 await expect(createStackReview('owner', {...draft, fields: ['description'], camera: null}, ids)).rejects.toThrow();
 await expect(createStackReview('owner', draft, Array.from({length: 51}, (_, index) => `00000000-0000-4000-8000-${String(index).padStart(12, '0')}`))).rejects.toThrow();
 expect(request).not.toHaveBeenCalled();
});

it('allows the stack preview its bounded thirty-second metadata read deadline', async () => {
 vi.useFakeTimers();
 try {
  const {createWritePreview} = await import('./writePreviewApi'); const {stackPreview} = await import('./testing/writePreview');
  const draft = stagedPreviewDraft(); const preview = stackPreview(); const review = stackReview(preview.plan.manifest.targets.map(target => target.assetId));
  let release: (value: Response) => void = () => undefined; let hasFailed = false;
  vi.stubGlobal('fetch', vi.fn(async () => new Promise<Response>(resolve => {release = resolve;})));
  const pending = createWritePreview('owner', draft, undefined, review).catch(() => {hasFailed = true; return null;});
  await vi.advanceTimersByTimeAsync(20_000);
  expect(hasFailed).toBe(false);
  release(new Response(JSON.stringify(preview)));
  await expect(pending).resolves.toEqual(preview);
 } finally {vi.useRealTimers();}
});
