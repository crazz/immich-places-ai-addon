export function makeStackFixture(original) {
	const assets = structuredClone(original);
	const primary = assets.find(asset => asset.id === '11111111-1111-4111-8111-111111111111');
	if (!primary) {throw new Error('Synthetic stack primary required');}
	const stack = {id: 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', primaryAssetId: primary.id, assets: [primary]};
	for (const digit of ['3', '4', '5', '6']) {
		const id = `${digit.repeat(8)}-${digit.repeat(4)}-4${digit.repeat(3)}-8${digit.repeat(3)}-${digit.repeat(12)}`;
		const sibling = {...structuredClone(primary), id, originalFileName: `${id}.png`, originalPath: `/synthetic/${id}.png`, exifInfo: {...primary.exifInfo, latitude: null, longitude: null, description: 'Preserved sibling description'}};
		assets.push(sibling);
		stack.assets.push(sibling);
	}
	for (const asset of stack.assets) {
		Object.assign(asset, {visibility: 'timeline', isTrashed: false, stack: {id: stack.id, primaryAssetId: primary.id, assetCount: stack.assets.length}});
	}
	return {assets, stack};
}
