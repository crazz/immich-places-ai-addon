import {afterEach, expect, it, vi} from 'vitest';

import {fetchProviders} from './providerApi';

const profile = {id: 'private', name: 'Private', baseURL: 'https://provider.example/v1', model: 'manual-model', enabled: true, revision: 1, hasSecret: true};

afterEach(() => vi.unstubAllGlobals());

it('rejects malformed execution readiness instead of authorizing a launch', async () => {
 const request = vi.fn<typeof fetch>();
 vi.stubGlobal('fetch', request);
 request.mockResolvedValueOnce(Response.json({enabled: true, items: [{...profile, executionReadiness: {status: 'ready', contextAllowed: 'yes', maxInputTokens: -1}}]}));
 await expect(fetchProviders()).rejects.toThrow('Invalid provider list');
});
