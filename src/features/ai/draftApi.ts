import {backendFetch, parseJSON} from '@/shared/services/backendApi.fetch';
import {getBackendBaseURL} from '@/utils/backendUrls';
import {isRecord} from '@/utils/typeGuards';

import {AIRequestError} from './aiRequest';
import {raceAbort, withCompleteOperationDeadline} from './completeOperation';
import {isDraft} from './draftTypes';

import type {TDraftEdit} from './draftEdit';
import type {TDraft} from './draftTypes';

export async function draftRequest<T>(path: string, method: 'GET' | 'POST' | 'PATCH', input: unknown, validate: (value: unknown) => value is T, signal?: AbortSignal, revision?: number): Promise<T> {
 return withCompleteOperationDeadline(30_000, signal, async operationSignal => raceAbort(operationSignal, (async () => {
  const headers = new Headers([['Content-Type', 'application/json']]);
  if (revision !== undefined) {headers.set('If-Match', `"${revision}"`);}
  const response = await backendFetch(`${getBackendBaseURL()}${path}`, {method, cache: 'no-store', headers, body: method === 'GET' ? undefined : JSON.stringify(input)}, {signal: operationSignal, timeoutMs: 30_000});
  if (!response.ok) {
   let failure: unknown; try {failure = await response.json();} catch {failure = null;}
   throw new AIRequestError(isRecord(failure) && typeof failure.code === 'string' ? failure.code : 'REQUEST_FAILED', 'Draft request failed. Reload saved state before retrying.');
  }
  return parseJSON(response, validate, 'Invalid draft response');
 })()));
}
export async function acceptDraft(analysisId: string, candidateId: string | null, signal?: AbortSignal): Promise<TDraft> {
 return draftRequest(`/ai/results/${encodeURIComponent(analysisId)}/draft`, 'POST', {candidateId}, isDraft, signal);
}
export async function fetchDraft(id: string, signal?: AbortSignal): Promise<TDraft> {return draftRequest(`/ai/drafts/${encodeURIComponent(id)}`, 'GET', undefined, isDraft, signal);}

export async function saveDraft(draft: TDraft, edit: TDraftEdit, signal?: AbortSignal): Promise<TDraft> {
 return draftRequest(`/ai/drafts/${encodeURIComponent(draft.id)}`, 'PATCH', edit, isDraft, signal, draft.revision);
}
