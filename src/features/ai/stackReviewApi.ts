import {backendFetch, parseJSON} from '@/shared/services/backendApi.fetch';
import {getBackendBaseURL} from '@/utils/backendUrls';

import {raceAbort, withCompleteOperationDeadline} from './completeOperation';
import {isDraftPoint} from './draftTypes';
import {isUUID} from './selectionTypes';
import {isStackReview} from './stackReviewTypes';

import type {TDraft} from './draftTypes';
import type {TStackReview} from './stackReviewTypes';

export async function createStackReview(owner: string, draft: TDraft, targetIds: string[], signal?: AbortSignal): Promise<TStackReview> {
 const selected = [...new Set(targetIds.length ? targetIds : [draft.assetId])].sort();
 if (!selected.includes(draft.assetId) || selected.length > 50 || !selected.every(isUUID) || (selected.length > 1 && (!draft.fields.includes('gps') || !isDraftPoint(draft.camera)))) {throw new Error('Invalid stack selection');}
 return withCompleteOperationDeadline(30_000, signal, async operationSignal => raceAbort(operationSignal, (async () => {
  const response = await backendFetch(`${getBackendBaseURL()}/ai/stack-reviews`, {method: 'POST', cache: 'no-store', headers: new Headers([['Content-Type', 'application/json']]), body: JSON.stringify({draftId: draft.id, draftRevision: draft.revision, targetIds: selected})}, {signal: operationSignal, timeoutMs: 30_000});
  if (!response.ok) {throw new Error('Stack review unavailable');}
  const value = await parseJSON(response, isStackReview, 'Invalid stack review');
  const ids = selected;
  if (Date.parse(value.expiresAt) <= Date.now() || value.owner !== owner || value.draftId !== draft.id || value.draftRevision !== draft.revision || value.analyzedId !== draft.assetId || value.imageIdentity !== draft.baseline.imageIdentity || value.targets.length !== ids.length || !value.targets.every(target => ids.includes(target.assetId))) {throw new Error('Stack review changed');}
  return value;
 })()));
}
