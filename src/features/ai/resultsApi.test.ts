import {afterEach, expect, it, vi} from 'vitest';

import {fetchResult, fetchResults} from './resultsApi';
import {resultDetail} from './testing/resultDetail';
import {resultEntry} from './testing/results';

afterEach(() => vi.unstubAllGlobals());

it('reads the exact bounded filters and opaque cursor without mutations or automatic retries', async () => {
	const page = {items: [resultEntry()], nextCursor: 'opaque'};
	const request = vi.fn<typeof fetch>().mockResolvedValue(Response.json(page));
	vi.stubGlobal('fetch', request);
	await expect(fetchResults({state: 'succeeded', startDate: '2026-09-01', cursor: 'opaque/+='})).resolves.toEqual(page);
	const url = new URL(String(request.mock.calls[0][0]), 'http://localhost');
	expect(url.pathname).toContain('/ai/results');
	expect(Object.fromEntries(url.searchParams)).toEqual({state: 'succeeded', startDate: '2026-09-01', cursor: 'opaque/+='});
	expect(request.mock.calls[0][1]?.method).toBe('GET');
	expect(request.mock.calls[0][1]?.body).toBeUndefined();
	expect(request.mock.calls[0][1]?.cache).toBe('no-store');
	expect(request).toHaveBeenCalledTimes(1);
});

it('resolves exact successful or failed item detail with validated identities', async () => {
	const detail = resultDetail();
	const failed = {...detail, entry: resultEntry({executionState: 'failed', analysisId: null, proposalOutcome: null}), proposal: null};
	const request = vi.fn<typeof fetch>().mockResolvedValueOnce(Response.json(detail)).mockResolvedValueOnce(Response.json(failed));
	vi.stubGlobal('fetch', request);
	await expect(fetchResult({analysisId: detail.entry.analysisId!})).resolves.toEqual(detail);
	await expect(fetchResult({jobId: failed.entry.jobId, itemId: failed.entry.id})).resolves.toEqual(failed);
	expect(String(request.mock.calls[0][0])).toContain(`/ai/results/${detail.entry.analysisId}`);
	expect(String(request.mock.calls[1][0])).toContain(`/ai/jobs/${failed.entry.jobId}/items/${failed.entry.id}/result`);
	await expect(fetchResult({analysisId: '../providers'})).rejects.toThrow();
	expect(request).toHaveBeenCalledTimes(2);
});
