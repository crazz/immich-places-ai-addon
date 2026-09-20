import {expect, it} from 'vitest';

import {isSelectionPreview} from './selectionTypes';
import {aiPreview} from './testing/fixtures';

it('rejects malformed authoritative scope fields before rendering or consent', () => {
	const preview = aiPreview();
	expect(isSelectionPreview(preview)).toBe(true);
	for (const field of ['gpsFilter', 'hiddenFilter', 'albumID', 'folderPath', 'tagID', 'startDate', 'endDate']) {
		expect(isSelectionPreview({...preview, scope: {...preview.scope, [field]: {unexpected: 'object'}}})).toBe(false);
	}
	expect(isSelectionPreview({...preview, scope: {view: 'all'}})).toBe(false);
});
