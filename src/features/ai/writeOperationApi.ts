import {backendFetch, parseJSON} from '@/shared/services/backendApi.fetch';
import {getBackendBaseURL} from '@/utils/backendUrls';
import {isRecord} from '@/utils/typeGuards';

import {AIRequestError} from './aiRequest';
import {raceAbort, withCompleteOperationDeadline} from './completeOperation';
import {isWriteHistory, isWriteOperation} from './writeOperationTypes';

import type {TDraft} from './draftTypes';
import type {TWriteHistory, TWriteOperation} from './writeOperationTypes';

async function request<T>(path: string, input: unknown, validate: (value: unknown) => value is T, signal?: AbortSignal): Promise<T> {
 return withCompleteOperationDeadline(15_000, signal, async operationSignal => raceAbort(operationSignal, (async () => {
  const response = await backendFetch(`${getBackendBaseURL()}/ai/write-operations${path}`, {method: input === undefined ? 'GET' : 'POST', redirect: 'error', cache: 'no-store', headers: new Headers([['Content-Type', 'application/json']]), body: input === undefined ? undefined : JSON.stringify(input)}, {signal: operationSignal, timeoutMs: 15_000});
  if (!response.ok) {
   let value: unknown; try {value = await response.json();} catch {value = null;}
   throw new AIRequestError(isRecord(value) && typeof value.code === 'string' ? value.code : 'WRITE_UNAVAILABLE', 'GPS operation unavailable.');
  }
  return parseJSON(response, validate, 'Invalid GPS operation response');
 })()));
}

export async function writeOperationRequest(owner: string, draft: TDraft, path: string, input: unknown, signal?: AbortSignal): Promise<TWriteOperation> {
 const value = await request(path, input, isWriteOperation, signal);
 if (value.plan.owner !== owner || value.plan.draftId !== draft.id || value.plan.analysisId !== draft.analysisId || value.plan.targetId !== draft.assetId) {throw new Error('Operation scope changed');}
 return value;
}

export async function fetchWriteHistory(draftID: string, cursor: string, signal?: AbortSignal): Promise<TWriteHistory> {
 return request(`?draftId=${encodeURIComponent(draftID)}${cursor ? `&before=${encodeURIComponent(cursor)}` : ''}`, undefined, isWriteHistory, signal);
}
