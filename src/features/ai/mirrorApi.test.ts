import {afterEach, expect, it, vi} from 'vitest';

import {mirrorPreview, mirroredDraft} from './testing/mirror';
import {savedPreview} from './testing/writePreview';
import {createWritePreview} from './writePreviewApi';

afterEach(() => vi.unstubAllGlobals());

it('binds each selected mirror value and revision without silent downgrade or inferred disclosure', async () => {
 const draft = mirroredDraft(); const preview = mirrorPreview(draft);
 const request = vi.fn(async () => new Response(JSON.stringify(preview))); vi.stubGlobal('fetch', request);
 await expect(createWritePreview('owner', draft)).rejects.toThrow(); expect(request).not.toHaveBeenCalled();
 await expect(createWritePreview('owner', draft, undefined, undefined, 'asset-readers-v1')).resolves.toEqual(preview);
 for (const value of [savedPreview(), {...preview, plan: {...preview.plan, mirror: {...preview.plan.mirror, value: {...preview.plan.mirror.value, descriptions: {uk: 'Replaced text'}}}}}, {...preview, plan: {...preview.plan, mirror: {...preview.plan.mirror, value: {...preview.plan.mirror.value, direction: {heading: 42, method: 'visual_estimate', uncertainty: null}}}}}]) {
  request.mockResolvedValueOnce(new Response(JSON.stringify(value)));
  await expect(createWritePreview('owner', draft, undefined, undefined, 'asset-readers-v1')).rejects.toThrow();
 }
 request.mockResolvedValueOnce(new Response(JSON.stringify(preview)));
 await expect(createWritePreview('owner', {...draft, mirror: undefined})).rejects.toThrow();
});

it('rejects foreign metadata recovery identity or responses without hidden mutation retries', async () => {
 const {mirrorOperationRequest} = await import('./mirrorApi'); const {mirrorOperation} = await import('./testing/mirror'); const draft = mirroredDraft(); const op = mirrorOperation(); const identity = {operationId: op.id, assetId: draft.assetId, generation: 3};
 const request = vi.fn<(path: string, init: RequestInit) => Promise<Response>>(async () => new Response(JSON.stringify(op))); vi.stubGlobal('fetch', request);
 for (const invalid of [{...identity, operationId: '../foreign'}, {...identity, assetId: '55555555-5555-4555-8555-555555555555'}, {...identity, generation: 0}, {...identity, generation: 1.5}]) {await expect(mirrorOperationRequest('owner', draft, invalid, 'retry')).rejects.toThrow();}
 expect(request).not.toHaveBeenCalled();
 await expect(mirrorOperationRequest('owner', draft, identity, 'reconcile')).resolves.toEqual(op); expect(request).toHaveBeenCalledTimes(1); expect(JSON.parse(String(request.mock.calls[0][1].body))).toEqual({generation: 3});
 for (const changed of [{...op, id: '55555555-5555-4555-8555-555555555555'}, {...op, plan: {...op.plan, owner: 'other'}}, {...op, mirror: {...op.mirror, assetId: '55555555-5555-4555-8555-555555555555'}}]) {
  request.mockResolvedValueOnce(new Response(JSON.stringify(changed))); await expect(mirrorOperationRequest('owner', draft, identity, 'retry')).rejects.toThrow();
 }
 request.mockRejectedValueOnce(new Error('lost response')); await expect(mirrorOperationRequest('owner', draft, identity, 'retry')).rejects.toThrow(); expect(request).toHaveBeenCalledTimes(5);
});
