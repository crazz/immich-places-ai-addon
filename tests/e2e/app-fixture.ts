import {randomUUID} from 'node:crypto';

import {test as base, expect} from '@playwright/test';

import type {APIRequestContext, Locator, Page} from '@playwright/test';

export const MATCHED_ID = '11111111-1111-4111-8111-111111111111';
export const UNMATCHED_ID = '22222222-2222-4222-8222-222222222222';
const PASSWORD = 'synthetic-password';
const TILE = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aD1sAAAAASUVORK5CYII=', 'base64');

type TAccount = {email: string; key: string};
type TWrite = {ids: string[]; latitude: number; longitude: number};
type TFixtureState = {writes: TWrite[]; errors: string[]};
type TProviderRequest = {
	path: string;
	authorization: string;
	model: string;
	hasImage: boolean;
	stream: boolean;
};
type TProviderState = {requests: TProviderRequest[]; errors: string[]};

export async function fixtureState(request: APIRequestContext, key: string): Promise<TFixtureState> {
	const response = await request.get(`http://127.0.0.1:8090/__state/${key}`);
	expect(response.ok()).toBe(true);
	return response.json();
}

export async function providerState(request: APIRequestContext): Promise<TProviderState> {
	const response = await request.get('http://127.0.0.1:8090/__provider_state');
	expect(response.ok()).toBe(true);
	return response.json();
}

export function photo(page: Page, assetID: string): Locator {
	return page.locator('[draggable="true"]').filter({has: page.locator(`img[src*="${assetID}"]`)});
}

export async function connectAccount(page: Page, account: TAccount): Promise<void> {
	await page.goto('/');
	await page.getByRole('button', {name: 'Create an account'}).click();
	await page.getByPlaceholder('Email', {exact: true}).fill(account.email);
	await page.getByPlaceholder('Password', {exact: true}).fill(PASSWORD);
	await page.getByPlaceholder('Confirm password', {exact: true}).fill(PASSWORD);
	await page.getByRole('button', {name: 'Create account', exact: true}).click();
	await page.getByPlaceholder('Immich API Key', {exact: true}).fill(account.key);
	await page.getByRole('button', {name: 'Save API Key', exact: true}).click();
	await page.getByRole('button', {name: 'Missing location'}).click();
	await page.getByRole('button', {name: 'Timeline', exact: true}).click();
	await expect(photo(page, MATCHED_ID)).toBeVisible();
	await expect(photo(page, UNMATCHED_ID)).toBeVisible();
	await expect(photo(page, MATCHED_ID).locator('img')).toHaveJSProperty('naturalWidth', 1);
	await expect(photo(page, UNMATCHED_ID).locator('img')).toHaveJSProperty('naturalWidth', 1);
}

export async function login(page: Page, email: string): Promise<void> {
	await page.getByPlaceholder('Email', {exact: true}).fill(email);
	await page.getByPlaceholder('Password', {exact: true}).fill(PASSWORD);
	await page.getByRole('button', {name: 'Sign in', exact: true}).click();
	await page.getByRole('button', {name: 'Missing location'}).click();
	await page.getByRole('button', {name: 'Timeline', exact: true}).click();
	await expect(photo(page, MATCHED_ID)).toBeVisible();
}

export async function expectPersisted(page: Page, latitude: number, longitude: number): Promise<void> {
	await page.reload();
	const response = await page.request.get('/api/backend/assets?pageSize=100&gpsFilter=all');
	expect(response.ok()).toBe(true);
	const data = await response.json();
	expect(data.items).toEqual(expect.arrayContaining([
		expect.objectContaining({immichID: MATCHED_ID, latitude, longitude}),
		expect.objectContaining({immichID: UNMATCHED_ID, latitude: null, longitude: null})
	]));
}

export const test = base.extend<{account: TAccount}>({
	account: async ({page}, provideAccount, testInfo) => {
		const key = `smoke-${randomUUID()}`;
		const blocked: string[] = [];
		const pageErrors: string[] = [];
		page.on('pageerror', error => pageErrors.push(error.message));
		await page.route('**/*', async route => {
			const url = new URL(route.request().url());
			if (url.origin === 'http://127.0.0.1:3080') {
				await route.continue();
			} else if (/^[abc]\.basemaps\.cartocdn\.com$/.test(url.hostname)) {
				await route.fulfill({contentType: 'image/png', body: TILE});
			} else {
				blocked.push(url.origin);
				await route.abort('blockedbyclient');
			}
		});
		await provideAccount({email: `${key}@example.test`, key});
		const state = await fixtureState(page.request, key);
		await testInfo.attach('synthetic-immich-state', {body: JSON.stringify(state, null, 2), contentType: 'application/json'});
		expect(blocked, 'Unexpected browser egress').toEqual([]);
		expect(pageErrors, 'Uncaught browser errors').toEqual([]);
		expect(state.errors, 'Unexpected backend/fake requests').toEqual([]);
	}
});

export {expect};
