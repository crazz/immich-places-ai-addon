import {afterEach, expect, it, vi} from 'vitest';

import {previewSelection} from './selectionApi';

const scope = {view: 'all' as const, gpsFilter: 'no-gps', hiddenFilter: 'visible'};
const preview = {mode: 'all-matching', scope, snapshotID: '11111111-1111-4111-8111-111111111111', policyVersion: 'selection-v1', assetIDs: ['aaaaaaaa-0000-4000-8000-000000000001'], requestedCount: 2, uniqueCount: 2, duplicateCount: 0, eligibleCount: 1, excludedCount: 1, matchedCount: 2, exclusionCounts: {hidden: 1}, createdAt: '2026-09-20T12:00:00Z', expiresAt: '2026-09-20T12:15:00Z'};

afterEach(() => vi.unstubAllGlobals());

it('previews the exact server scope without turning all matching into page IDs or retrying a mutation', async () => {
 const request = vi.fn<typeof fetch>().mockResolvedValueOnce(Response.json(preview));
 vi.stubGlobal('fetch', request);
 await expect(previewSelection({mode: 'all-matching', scope})).resolves.toEqual(preview);
 expect(request).toHaveBeenCalledTimes(1);
 expect(String(request.mock.calls[0][0])).toContain('/ai/selection-preview');
 expect(JSON.parse(String(request.mock.calls[0][1]?.body))).toEqual({mode: 'all-matching', scope});
 expect(request.mock.calls[0][1]?.method).toBe('POST');
});

it('reports authoritative over-limit counts without exposing arbitrary server text or retrying', async () => {
 const request = vi.fn<typeof fetch>().mockResolvedValueOnce(Response.json({code: 'SELECTION_LIMIT_EXCEEDED', snapshotID: null, matchedCount: 510, eligibleCount: 501, excludedCount: 9, exclusionCounts: {hidden: 9}, message: 'Untrusted upstream detail'}, {status: 413}));
 vi.stubGlobal('fetch', request);
 await expect(previewSelection({mode: 'all-matching', scope})).rejects.toThrow('Matched: 510 · Eligible: 501 · Excluded: 9');
 expect(request).toHaveBeenCalledTimes(1);
});
