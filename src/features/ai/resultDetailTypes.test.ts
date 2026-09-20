import {expect, it} from 'vitest';

import {isResultDetail} from './resultDetailTypes';
import {resultDetail} from './testing/resultDetail';
import {resultEntry} from './testing/results';


const detailFixture = resultDetail();

it('validates immutable detail and separates failed entries from successful proposals', () => {
	expect(isResultDetail(detailFixture)).toBe(true);
	for (const patch of [{proposal: null}, {proposal: {...detailFixture.proposal, observations: [{id: 'obs', kind: 'visual', text: 7}]}}, {provenance: {...detailFixture.provenance, Model: 'different'}}, {proposal: {...detailFixture.proposal, outcome: 'located'}}, {provenance: {...detailFixture.provenance, Languages: ['uk']}}]) {
		expect(isResultDetail({...detailFixture, ...patch})).toBe(false);
	}
	expect(isResultDetail({...detailFixture, entry: resultEntry({executionState: 'failed', analysisId: null, proposalOutcome: null}), proposal: null})).toBe(true);
});
