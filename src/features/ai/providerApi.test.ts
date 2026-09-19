import {afterEach, expect, it, vi} from 'vitest';

import {CAPABILITY_TEST_TIMEOUT_MS, fetchProviders, saveProvider, testProvider} from './providerApi';

const profile = {id: 'private', name: 'Private', baseURL: 'https://provider.example/v1', model: 'manual-model', enabled: true, revision: 1, hasSecret: true};

const report = {
	attemptID: 'attempt-1',
	profileID: profile.id,
	revision: 1,
	protocolVersion: 'capability-v1',
	policyFingerprint: 'fp',
	lifecycle: 'completed',
	startedAt: '2026-09-19T12:00:00.000Z',
	deadlineAt: '2026-09-19T12:02:00.000Z',
	completedAt: '2026-09-19T12:00:05.000Z',
	requestedModel: profile.model,
	observations: {
		image: {status: 'supported'},
		json: {status: 'supported'},
		strict: {status: 'supported'}
	},
	compatibility: 'strict-schema sample compatible',
	applicable: true
};

afterEach(() => {
	vi.unstubAllGlobals();
	vi.useRealTimers();
});

it('validates redacted responses and sends exact revision-bound saves without retries', async () => {
	const request = vi.fn<typeof fetch>();
	vi.stubGlobal('fetch', request);
	request.mockResolvedValueOnce(Response.json({enabled: true, items: [profile]}));
	await expect(fetchProviders()).resolves.toEqual({enabled: true, items: [profile]});
	request.mockResolvedValueOnce(Response.json({...profile, revision: 2}));
	await expect(saveProvider({name: 'Private', baseURL: profile.baseURL, model: 'manual-model', enabled: false}, profile)).resolves.toMatchObject({revision: 2});
	const [, options] = request.mock.calls[1];
	expect(options?.method).toBe('PUT');
	expect(JSON.parse(String(options?.body))).toEqual({name: 'Private', baseURL: profile.baseURL, model: 'manual-model', enabled: false, expectedRevision: 1});
	request.mockResolvedValueOnce(Response.json(profile));
	await saveProvider({name: 'Private', baseURL: profile.baseURL, model: profile.model, enabled: true, secret: 'synthetic-secret'});
	expect(request.mock.calls[2][1]?.method).toBe('POST');
	expect(JSON.parse(String(request.mock.calls[2][1]?.body))).not.toHaveProperty('expectedRevision');
	request.mockResolvedValueOnce(Response.json({code: 'PROVIDER_CONFLICT', message: 'Reload before editing', retryable: false, requestID: 'request'}, {status: 409}));
	await expect(saveProvider({...profile, secret: ''}, profile)).rejects.toMatchObject({message: 'Reload before editing', code: 'PROVIDER_CONFLICT'});
	expect(request).toHaveBeenCalledTimes(4);
	request.mockResolvedValueOnce(Response.json({enabled: true, items: [{...profile, revision: 'wrong'}]}));
	await expect(fetchProviders()).rejects.toThrow('Invalid provider list');
	request.mockResolvedValueOnce(new Response('invalid', {status: 500}));
	await expect(fetchProviders()).rejects.toThrow('Provider settings request failed (500)');
	request.mockResolvedValueOnce(Response.json({message: ''}, {status: 400}));
	await expect(fetchProviders()).rejects.toThrow('Provider settings request failed (400)');
	request.mockResolvedValueOnce(Response.json({enabled: false, items: []}));
	await expect(fetchProviders()).resolves.toEqual({enabled: false, items: []});
	request.mockResolvedValueOnce(Response.json(null));
	await expect(saveProvider(profile)).rejects.toThrow('Invalid provider response');
});

it('posts a revision-bound capability test with a complete-operation deadline', async () => {
	const request = vi.fn<typeof fetch>();
	vi.stubGlobal('fetch', request);
	const controller = new AbortController();
	request.mockResolvedValueOnce(Response.json(report));
	await expect(testProvider(profile, controller.signal)).resolves.toEqual(report);
	expect(request).toHaveBeenCalledTimes(1);
	const [url, options] = request.mock.calls[0];
	expect(String(url)).toContain(`/ai/providers/${profile.id}/test`);
	expect(options?.method).toBe('POST');
	expect(JSON.parse(String(options?.body))).toEqual({expectedRevision: 1});
	expect(JSON.stringify(options?.body)).not.toContain('secret');
	expect(options?.signal).toBeInstanceOf(AbortSignal);
	request.mockResolvedValueOnce(Response.json({...report, observations: {image: {status: 'nope'}}}));
	await expect(testProvider(profile)).rejects.toThrow('Invalid capability report');
	request.mockResolvedValueOnce(Response.json({code: 'PROVIDER_BUSY', message: 'busy', retryable: false, requestID: 'r'}, {status: 409}));
	await expect(testProvider(profile)).rejects.toMatchObject({code: 'PROVIDER_BUSY', message: 'busy'});
	expect(CAPABILITY_TEST_TIMEOUT_MS).toBe(130_000);
	request.mockResolvedValueOnce(Response.json({enabled: true, items: [{...profile, capabilityReport: report}]}));
	await expect(fetchProviders()).resolves.toMatchObject({items: [{capabilityReport: report}]});
	request.mockResolvedValueOnce(Response.json({enabled: true, items: [{...profile, capabilityReport: {bad: true}}]}));
	await expect(fetchProviders()).rejects.toThrow('Invalid provider list');
});

it('aborts a capability test when the complete-operation deadline elapses after response headers', async () => {
	vi.useFakeTimers();
	const request = vi.fn<typeof fetch>();
	vi.stubGlobal('fetch', request);
	const headersArrived = Promise.withResolvers<void>();
	request.mockImplementationOnce(async () => {
		headersArrived.resolve();
		const headers = new Headers();
		headers.set('Content-Type', 'application/json');
		return new Response(new ReadableStream({
			start() {
				/* body intentionally never enqueues — deadline must win after headers */
			}
		}), {status: 200, headers});
	});
	const pending = testProvider(profile);
	const assertion = expect(pending).rejects.toMatchObject({name: 'AbortError'});
	await headersArrived.promise;
	await vi.advanceTimersByTimeAsync(CAPABILITY_TEST_TIMEOUT_MS);
	await assertion;
	vi.useRealTimers();
});
