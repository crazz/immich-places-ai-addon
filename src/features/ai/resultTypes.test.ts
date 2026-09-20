import {expect, it} from 'vitest';

import {isResultPage} from './resultTypes';
import {resultEntry} from './testing/results';

it('accepts bounded history and rejects inconsistent terminal identities and states', () => {
	const entry = resultEntry();
	expect(isResultPage({items: [entry]})).toBe(true);
	for (const patch of [{id: 'foreign'}, {sourceAvailable: 'yes'}, {terminalAt: 'invalid'}, {reviewState: 'approved'}, {writeState: 'succeeded'}, {executionState: 'failed'}, {proposalOutcome: null}, {label: {html: 'text'}}, {captureDay: 0}]) {
		expect(isResultPage({items: [{...entry, ...patch}]})).toBe(false);
	}
	expect(isResultPage({items: [entry, entry]})).toBe(false);
	expect(isResultPage({items: Array.from({length: 101}, () => entry)})).toBe(false);
	expect(isResultPage({items: [], nextCursor: 1})).toBe(false);
});
