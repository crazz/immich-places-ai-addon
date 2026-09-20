import {expect, it} from 'vitest';

import {isSelectionPreview} from './selectionTypes';
import {aiPreview} from './testing/fixtures';

it('bounds optional displayed context and rejects foreign assets and malformed values', () => {
	const base = aiPreview();
	const valid = {albumLabel: null, captureTimes: {[base.assetIDs[0]]: '2009-04-01T12:30:00'}};
	expect(isSelectionPreview({...base, contextPreview: valid})).toBe(true);
	for (const contextPreview of [
		null,
		{},
		{...valid, captureTimes: []},
		{...valid, albumLabel: 9},
		{...valid, albumLabel: 'x'.repeat(257)},
		{...valid, captureTimes: {foreign: 'date'}},
		{...valid, captureTimes: {[base.assetIDs[0]]: 'é'.repeat(65)}}
	]) {
		expect(isSelectionPreview({...base, contextPreview})).toBe(false);
	}
});
