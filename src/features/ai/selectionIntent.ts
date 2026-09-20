import type {TSelectionInput, TSelectionScope} from './selectionTypes';

export type TBatchIntention = 'selected' | 'page' | 'matching';
export type TBatchScope = {scope: TSelectionScope; selected: string[]; page: string[]; blockedReason: string; revisionKey?: string};

export function selectionIntent(kind: TBatchIntention, input: TBatchScope): TSelectionInput {
	if (input.blockedReason) {
		throw new Error(input.blockedReason);
	}
	if (kind === 'matching') {
		return {mode: 'all-matching', scope: input.scope};
	}
	const assetIDs = kind === 'selected' ? input.selected : input.page;
	if (assetIDs.length === 0) {
		throw new Error('Select at least one asset or choose a nonempty page.');
	}
	return {mode: 'explicit', scope: input.scope, assetIDs: [...assetIDs]};
}
