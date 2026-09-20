import {afterEach, expect, it, vi} from 'vitest';

import {submitJob} from './jobApi';

import type {TJobConfiguration} from './jobTypes';

const configuration: TJobConfiguration = {selectionToken: '11111111-1111-4111-8111-111111111111', profileId: 'p', revision: 1, mode: 'visual', format: 'strict', allowJson: false, languages: ['en'], primaryLanguage: 'en', policyId: 'a'.repeat(64), limits: {maxCalls: 1, maxTokens: 104000, outputTokens: 4000}};
const admission = {configuration, consent: {version: 'image-consent-v1', image: true, configuration}, idempotencyKey: 'stable-request-key'};
const progress = {id: '22222222-2222-4222-8222-222222222222', configuration, createdAt: 1789919280000000000, canceled: false, blocked: false, counts: {queued: 1, total: 1}, items: [{id: '33333333-3333-4333-8333-333333333333', assetId: 'aaaaaaaa-0000-4000-8000-000000000001', state: 'queued', attempts: 0, calls: 0, resultId: null}], usage: {calls: 0, reservedTokens: 0, reportedStatus: 'unknown', costStatus: 'unknown', inputReported: null, outputReported: null, totalReported: null, estimatedMicros: null}};

afterEach(() => vi.unstubAllGlobals());

it('reconciles an explicitly repeated submission with the same key and no automatic retry', async () => {
 const request = vi.fn<typeof fetch>().mockRejectedValueOnce(new Error('acknowledgement lost')).mockResolvedValueOnce(Response.json(progress));
 vi.stubGlobal('fetch', request);
 await expect(submitJob(admission)).rejects.toThrow('acknowledgement lost');
 expect(request).toHaveBeenCalledTimes(1);
 await expect(submitJob(admission)).resolves.toEqual(progress);
 expect(request).toHaveBeenCalledTimes(2);
 for (const [url, options] of request.mock.calls) {
  expect(String(url)).toContain('/ai/jobs');
  expect(options?.method).toBe('POST');
  expect(JSON.parse(String(options?.body))).toEqual(admission);
 }
});

it('reads durable progress and pages and cancels once without submitting another job', async () => {
 const {fetchJob, fetchJobs, cancelJob} = await import('./jobApi');
 const request = vi.fn<typeof fetch>().mockResolvedValueOnce(Response.json(progress)).mockResolvedValueOnce(Response.json({items: [{...progress, items: null}], nextCursor: null})).mockResolvedValueOnce(Response.json({canceled: true}));
 vi.stubGlobal('fetch', request);
 await expect(fetchJob(progress.id)).resolves.toEqual(progress);
 await expect(fetchJobs()).resolves.toMatchObject({items: [{id: progress.id}]});
 await expect(cancelJob(progress.id)).resolves.toEqual({canceled: true});
 expect(request.mock.calls.map(call => call[1]?.method)).toEqual(['GET', 'GET', 'POST']);
 expect(String(request.mock.calls[2][0])).toContain(`/ai/jobs/${progress.id}/cancel`);
 expect(JSON.parse(String(request.mock.calls[2][1]?.body))).toEqual({});
});
