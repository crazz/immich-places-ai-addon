import {aiRequest} from './aiRequest';
import {isSelectionPreview} from './selectionTypes';

import type {TSelectionInput, TSelectionPreview} from './selectionTypes';

export async function previewSelection(input: TSelectionInput, signal?: AbortSignal): Promise<TSelectionPreview> {
	return aiRequest('/ai/selection-preview', 'POST', input, isSelectionPreview, signal);
}
