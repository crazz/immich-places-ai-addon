import {MATCHED_ID, connectAccount, expect, expectPersisted, fixtureState, photo, test} from './app-fixture';

test('manual preview and cancellation do not write; confirmation changes only the selected photo', async ({page, account}) => {
	await connectAccount(page, account);
	await photo(page, MATCHED_ID).click();
	await page.getByPlaceholder('Search location...').fill('52.25, 21.05');
	await expect(page.getByText('1 edited image', {exact: true})).toBeVisible();
	expect((await fixtureState(page.request, account.key)).writes).toEqual([]);
	await page.getByRole('button', {name: 'Cancel', exact: true}).click();
	await expect(page.getByRole('button', {name: 'Save Location', exact: true})).not.toBeVisible();
	expect((await fixtureState(page.request, account.key)).writes).toEqual([]);
	await page.getByPlaceholder('Search location...').fill('52.26, 21.06');
	await page.getByRole('button', {name: 'Save Location', exact: true}).click();
	await expect.poll(async () => (await fixtureState(page.request, account.key)).writes).toEqual([{ids: [MATCHED_ID], latitude: 52.26, longitude: 21.06}]);
	await expect(page.getByRole('button', {name: 'Saving...', exact: true})).not.toBeVisible();
	await expectPersisted(page, 52.26, 21.06);
});
