import {MATCHED_ID, connectAccount, expect, expectPersisted, fixtureState, test} from './app-fixture';

const GPX = `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" creator="synthetic-smoke"><trk><name>Synthetic track</name><trkseg>
<trkpt lat="52" lon="21"><time>2026-08-01T10:00:00Z</time></trkpt>
<trkpt lat="52" lon="21"><time>2026-08-01T10:10:00Z</time></trkpt>
</trkseg></trk></gpx>`;

test('malformed GPX and valid preview do not write; confirmation changes only the matched photo', async ({page, account}) => {
	await connectAccount(page, account);
	await page.getByTitle(account.email, {exact: true}).click();
	await page.getByRole('button', {name: 'Import Tracks', exact: true}).click();
	const input = page.locator('input[type="file"]');
	const invalidPreview = page.waitForResponse(response => response.url().endsWith('/api/backend/gpx/preview'));
	await input.setInputFiles({name: 'broken.gpx', mimeType: 'application/gpx+xml', buffer: Buffer.from('<gpx><')});
	expect((await invalidPreview).status()).toBe(400);
	await expect(page.getByText(/broken\.gpx:/)).toBeVisible();
	expect((await fixtureState(page.request, account.key)).writes).toEqual([]);
	await input.setInputFiles({name: 'synthetic.gpx', mimeType: 'application/gpx+xml', buffer: Buffer.from(GPX)});
	await expect(page.getByText('1 edited image', {exact: true})).toBeVisible();
	expect((await fixtureState(page.request, account.key)).writes).toEqual([]);
	await page.getByRole('button', {name: 'Save Location', exact: true}).click();
	await expect.poll(async () => (await fixtureState(page.request, account.key)).writes).toEqual([{ids: [MATCHED_ID], latitude: 52, longitude: 21}]);
	await expect(page.getByRole('button', {name: 'Saving...', exact: true})).not.toBeVisible();
	await expectPersisted(page, 52, 21);
});
