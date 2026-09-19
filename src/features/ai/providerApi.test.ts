import {afterEach, expect, it, vi} from 'vitest';

import {fetchProviders, saveProvider} from './providerApi';

const profile = {id: 'private', name: 'Private', baseURL: 'https://provider.example/v1', model: 'manual-model', enabled: true, revision: 1, hasSecret: true};

afterEach(() => vi.unstubAllGlobals());

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
