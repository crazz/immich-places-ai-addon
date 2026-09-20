import {MATCHED_ID, expect, photo} from './app-fixture';

import type {Page} from '@playwright/test';

export async function launchReview(page: Page, model: string): Promise<void> {
 await page.getByRole('button', {name: 'Settings', exact: true}).click();
 await page.getByRole('button', {name: 'AI providers', exact: true}).click();
 await page.getByRole('button', {name: 'Create provider'}).click();
 await page.getByLabel('Name', {exact: true}).fill('Review synthetic');
 await page.getByLabel('API base URL', {exact: true}).fill('http://127.0.0.1:8090/v1');
 await page.getByLabel('Model', {exact: true}).fill(model);
 await page.getByLabel('API key', {exact: true}).fill('synthetic-review-secret');
 await page.getByRole('button', {name: 'Save provider'}).click();
 await expect(page.getByText('Saved revision 1', {exact: true})).toBeVisible();
 await page.getByRole('button', {name: 'Test provider'}).click();
 await expect(page.getByText(/Strict: supported/i)).toBeVisible({timeout: 45000});
 const me = await (await page.request.get('/api/backend/auth/me')).json();
 const profiles = await (await page.request.get('/api/backend/ai/providers')).json();
 const configured = await page.request.post('http://127.0.0.1:8091/configure', {data: {owner: me.user.ID, profile: profiles.items[0].id}});
 expect(configured.ok()).toBe(true);
 await page.getByRole('button', {name: 'Close dialog'}).click();
 await photo(page, MATCHED_ID).click();
 await page.getByRole('button', {name: 'AI Locate', exact: true}).click();
 await page.getByRole('button', {name: 'Preview selected assets'}).click();
 await expect(page.getByText(/Eligible: 1/)).toBeVisible();
 await page.getByLabel('Requested languages', {exact: true}).fill('en, uk');
 await page.getByRole('button', {name: 'Start analysis', exact: true}).click();
 await expect(page.getByText('Job complete', {exact: true})).toBeVisible({timeout: 15000});
}

export async function openReview(page: Page): Promise<void> {
 await page.getByRole('button', {name: 'AI Results', exact: true}).click();
 await page.getByRole('button', {name: /^Open result /}).first().click();
 await expect(page.getByRole('region', {name: 'Read-only proposal map'})).toBeVisible();
}
