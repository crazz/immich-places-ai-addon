import {connectAccount, expect, fixtureState, providerState, test} from './app-fixture';

test('manages a private profile through the real session, proxy and database without external writes', async ({page, account}) => {
	await connectAccount(page, account);
	await page.getByRole('button', {name: 'Settings', exact: true}).click();
	await page.getByRole('button', {name: 'AI providers', exact: true}).click();
	await page.getByRole('button', {name: 'Create provider'}).click();
	await page.getByLabel('Name', {exact: true}).fill('Local private profile');
	await page.getByLabel('API base URL', {exact: true}).fill('http://127.0.0.1:8090/never-call');
	await page.getByLabel('Model', {exact: true}).fill('manual-model');
	await page.getByLabel('API key', {exact: true}).fill('synthetic-provider-secret');
	await page.getByRole('button', {name: 'Save provider'}).click();
	await expect(page.getByText('Saved revision 1', {exact: true})).toBeVisible();
	await page.getByRole('button', {name: 'Edit Local private profile'}).click();
	await expect(page.getByLabel('API key', {exact: true})).toHaveValue('');
	await page.getByLabel('Enabled', {exact: true}).uncheck();
	await page.getByRole('button', {name: 'Save provider'}).click();
	await expect(page.getByText('Saved revision 2', {exact: true})).toBeVisible();
	await page.getByRole('button', {name: 'Close dialog'}).click();
	await page.reload();
	await page.getByRole('button', {name: 'Settings', exact: true}).click();
	await page.getByRole('button', {name: 'AI providers', exact: true}).click();
	await expect(page.getByText('manual-model · Disabled · Revision 2')).toBeVisible();
	const response = await page.request.get('/api/backend/ai/providers');
	expect(response.ok()).toBe(true);
	const profiles = await response.json();
	expect(profiles.items).toHaveLength(1);
	expect(profiles.items[0]).toMatchObject({revision: 2, enabled: false, hasSecret: true});
	expect(JSON.stringify(profiles)).not.toContain('synthetic-provider-secret');
	expect(await page.evaluate(() => JSON.stringify({...localStorage, ...sessionStorage}))).not.toContain('synthetic-provider-secret');
	expect((await fixtureState(page.request, account.key)).writes).toEqual([]);
	expect((await providerState(page.request)).requests).toEqual([]);
});

test('tests a saved provider through the real proxy and reloads persisted capability results', async ({page, account}) => {
	test.setTimeout(60_000);
	await connectAccount(page, account);
	await page.getByRole('button', {name: 'Settings', exact: true}).click();
	await page.getByRole('button', {name: 'AI providers', exact: true}).click();
	await page.getByRole('button', {name: 'Create provider'}).click();
	await page.getByLabel('Name', {exact: true}).fill('Loopback synthetic');
	await page.getByLabel('API base URL', {exact: true}).fill('http://127.0.0.1:8090/v1');
	await page.getByLabel('Model', {exact: true}).fill('manual-model');
	await page.getByLabel('API key', {exact: true}).fill('synthetic-provider-secret');
	await page.getByRole('button', {name: 'Save provider'}).click();
	await expect(page.getByText('Saved revision 1', {exact: true})).toBeVisible();
	await expect(page.getByText(/Synthetic-image test of this destination\/model/i)).toBeVisible();
	await expect(page.getByText(/provider usage/i)).toBeVisible();
	await page.getByRole('button', {name: 'Test provider'}).click();
	await expect(page.getByRole('status').filter({hasText: /Testing provider/i})).toBeVisible();
	await expect(page.getByText(/Image: supported/i)).toBeVisible({timeout: 45_000});
	await expect(page.getByText(/JSON: supported/i)).toBeVisible();
	await expect(page.getByText(/Strict: supported/i)).toBeVisible();
	const beforeReload = await providerState(page.request);
	expect(beforeReload.requests).toHaveLength(3);
	for (const entry of beforeReload.requests) {
		expect(entry.path).toBe('/v1/chat/completions');
		expect(entry.authorization).toBe('Bearer synthetic-provider-secret');
		expect(entry.model).toBe('manual-model');
		expect(entry.hasImage).toBe(true);
		expect(entry.stream).toBe(false);
	}
	await page.getByRole('button', {name: 'Reload profiles'}).click();
	await expect(page.getByText(/Image: supported/i)).toBeVisible();
	await expect(page.getByText(/JSON: supported/i)).toBeVisible();
	await expect(page.getByText(/Strict: supported/i)).toBeVisible();
	expect((await providerState(page.request)).requests).toHaveLength(3);
	expect(JSON.stringify(await (await page.request.get('/api/backend/ai/providers')).json())).not.toContain('synthetic-provider-secret');
	expect(await page.evaluate(() => JSON.stringify({...localStorage, ...sessionStorage}))).not.toContain('synthetic-provider-secret');
});

test('freezes an explicit selection through the protected proxy without external work', async ({page, account}) => {
	await connectAccount(page, account);
	const ids = ['11111111-1111-4111-8111-111111111111', '22222222-2222-4222-8222-222222222222'];
	const before = await fixtureState(page.request, account.key);
	expect(before.imageRequests).toBeGreaterThan(0);
	const providersBefore = await providerState(page.request);
	const response = await page.request.post('/api/backend/ai/selection-preview', {
		headers: {Origin: 'http://127.0.0.1:3080'},
		data: {mode: 'explicit', assetIDs: [...ids, ids[0]], scope: {view: 'all'}}
	});
	expect(response.status()).toBe(200);
	expect(response.headers()['cache-control']).toBe('no-store');
	const selection = await response.json();
	expect(selection).toMatchObject({assetIDs: ids, requestedCount: 3, uniqueCount: 2, duplicateCount: 1, eligibleCount: 2, excludedCount: 0});
	expect(typeof selection.snapshotID).toBe('string');
	const read = await page.request.get(`/api/backend/ai/selections/${selection.snapshotID}`);
	expect(read.status()).toBe(200);
	expect(await read.json()).toEqual(selection);
	const rejected = await page.request.post('/api/backend/ai/selection-preview', {
		headers: {Origin: 'https://wrong.example'},
		data: {mode: 'explicit', assetIDs: ids, scope: {view: 'all'}}
	});
	expect(rejected.status()).toBe(403);
	expect(await fixtureState(page.request, account.key)).toEqual(before);
	expect(await providerState(page.request)).toEqual(providersBefore);
	await expect(page.getByRole('button', {name: /Analyze selection/i})).toHaveCount(0);
});

test('freezes all matching photos through the proxy without external work', async ({page, account}) => {
	await connectAccount(page, account);
	const before = await fixtureState(page.request, account.key);
	const providersBefore = await providerState(page.request);
	const response = await page.request.post('/api/backend/ai/selection-preview', {
		headers: {Origin: 'http://127.0.0.1:3080'},
		data: {mode: 'all-matching', scope: {view: 'all'}}
	});
	expect(response.status()).toBe(200);
	expect(response.headers()['cache-control']).toBe('no-store');
	const selection = await response.json();
	expect(selection).toMatchObject({
		mode: 'all-matching',
matchedCount: 2,
requestedCount: 2,
uniqueCount: 2,
		duplicateCount: 0,
eligibleCount: 2,
excludedCount: 0,
exclusionCounts: {}
	});
	expect(selection.assetIDs).toEqual(['22222222-2222-4222-8222-222222222222', '11111111-1111-4111-8111-111111111111']);
	expect(selection).not.toHaveProperty('exclusions');
	expect(typeof selection.snapshotID).toBe('string');
	const read = await page.request.get(`/api/backend/ai/selections/${selection.snapshotID}`);
	expect(read.status()).toBe(200);
	expect(await read.json()).toEqual(selection);
	expect(await fixtureState(page.request, account.key)).toEqual(before);
	expect(await providerState(page.request)).toEqual(providersBefore);
});
