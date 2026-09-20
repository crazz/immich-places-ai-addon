import {connectAccount, expect, fixtureState, providerState, test} from './app-fixture';

test('exposes owner-scoped terminal history through the real application proxy', async ({page, account}) => {
	const before = (await providerState(page.request)).requests.length;
	await connectAccount(page, account);
	const response = await page.request.get('/api/backend/ai/results');
	expect(response.status()).toBe(200);
	expect(response.headers()['cache-control']).toBe('no-store');
	expect(await response.json()).toEqual({items: []});
	expect((await page.request.get('/api/backend/ai/results?limit=101')).status()).toBe(400);
	expect((await fixtureState(page.request, account.key)).writes).toEqual([]);
	expect((await providerState(page.request)).requests.length).toBe(before);
});
