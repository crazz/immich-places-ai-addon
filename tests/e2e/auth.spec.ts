import {MATCHED_ID, UNMATCHED_ID, connectAccount, expect, fixtureState, login, photo, test} from './app-fixture';

test('registers, connects Immich, browses and retains an authenticated session', async ({page, account}) => {
	await connectAccount(page, account);
	await page.reload();
	await expect(photo(page, MATCHED_ID)).toBeVisible();
	await expect(photo(page, UNMATCHED_ID)).toBeVisible();
	await page.getByTitle(account.email, {exact: true}).click();
	await page.getByRole('button', {name: 'Sign out', exact: true}).click();
	await expect(page.getByRole('heading', {name: 'Sign in', exact: true})).toBeVisible();
	expect((await page.request.get('/api/backend/assets')).status()).toBe(401);
	await login(page, account.email);
	expect((await fixtureState(page.request, account.key)).writes).toEqual([]);
});
