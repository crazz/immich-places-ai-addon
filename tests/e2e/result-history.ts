import {expect, providerState} from './app-fixture';

import type {Page} from '@playwright/test';

export async function reviewHistory(page: Page): Promise<void> {
 const before = (await providerState(page.request)).requests.length;
 await page.getByRole('button', {name: 'Close dialog'}).click();
 const entry = page.getByRole('button', {name: 'AI Results', exact: true});
 await entry.focus(); await entry.press('Enter');
 const history = page.getByRole('region', {name: 'Saved results'});
 await expect(history.getByText('Execution: succeeded · Proposal: unknown', {exact: true})).toHaveCount(3);
 await expect(history.getByRole('img', {name: 'Current source photo'})).toHaveCount(3);
 await page.getByText('Filter results', {exact: true}).click();
 await page.getByRole('combobox', {name: 'Outcome', exact: true}).selectOption('unknown');
 await page.getByRole('button', {name: 'Apply filters'}).click();
 const row = history.getByRole('button', {name: /^Open result /}).first();
 const rowName = await row.getAttribute('aria-label');
 await row.focus(); await row.press('Enter');
 await expect(page.getByRole('button', {name: 'Back to results'})).toBeFocused();
 await expect(page.getByText('Proposal: unknown', {exact: true})).toBeVisible();
 await expect(page.getByText('No location was established.', {exact: true})).toBeVisible();
 await page.getByText('Original run provenance', {exact: true}).click();
 await expect(page.getByText('context-assisted-v1', {exact: true})).toBeVisible();
 await page.screenshot({path: 'out/checks/ch14-apply/result-detail.png'});
 await page.getByRole('button', {name: 'Back to results'}).press('Enter');
 await expect(page.getByRole('button', {name: rowName!, exact: true})).toBeFocused();
 await expect(page.getByRole('combobox', {name: 'Outcome', exact: true})).toHaveValue('unknown');
 await page.getByRole('button', {name: rowName!, exact: true}).press('Enter');
 await page.reload();
 await expect(page.getByText('Proposal: unknown', {exact: true})).toBeVisible();
 await page.getByRole('button', {name: 'Close dialog'}).click();
 await expect(entry).toBeFocused();
 expect((await providerState(page.request)).requests.length).toBe(before);
}
