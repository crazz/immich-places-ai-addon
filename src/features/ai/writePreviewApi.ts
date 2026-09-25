import {backendFetch, parseJSON} from '@/shared/services/backendApi.fetch';
import {getBackendBaseURL} from '@/utils/backendUrls';
import {isRecord} from '@/utils/typeGuards';

import {AIRequestError} from './aiRequest';
import {raceAbort, withCompleteOperationDeadline} from './completeOperation';
import {isPreviewConflict, isWritePreview} from './writePreviewTypes';

import type {TDraft} from './draftTypes';
import type {TPreviewConflict, TWritePreview} from './writePreviewTypes';

export class PreviewRequestError extends AIRequestError {
 constructor(code: string, readonly conflict?: TPreviewConflict) {super(code, 'Preview unavailable.');}
}

async function previewRequest(path: string, method: 'GET' | 'POST', input: unknown, signal?: AbortSignal): Promise<TWritePreview> {
 return withCompleteOperationDeadline(15_000, signal, async operationSignal => raceAbort(operationSignal, (async () => {
  const response = await backendFetch(`${getBackendBaseURL()}${path}`, {method, cache: 'no-store', headers: new Headers([['Content-Type', 'application/json']]), body: method === 'POST' ? JSON.stringify(input) : undefined}, {signal: operationSignal, timeoutMs: 15_000});
  if (!response.ok) {
   let failure: unknown; try {failure = await response.json();} catch {failure = null;}
   throw new PreviewRequestError(isRecord(failure) && typeof failure.code === 'string' ? failure.code : 'REQUEST_FAILED', isRecord(failure) && failure.code === 'IMMICH_CONFLICT' && isPreviewConflict(failure.conflict) ? failure.conflict : undefined);
  }
  return parseJSON(response, isWritePreview, 'Invalid preview response');
 })()));
}

export async function createWritePreview(owner: string, draft: TDraft, signal?: AbortSignal): Promise<TWritePreview> {
 const value = await previewRequest('/ai/write-previews', 'POST', {draftId: draft.id, draftRevision: draft.revision}, signal);
 return checkScope(value, owner, draft);
}

export async function fetchWritePreview(owner: string, draft: TDraft, id: string, signal?: AbortSignal): Promise<TWritePreview> {
 const value = await previewRequest(`/ai/write-previews/${encodeURIComponent(id)}`, 'GET', undefined, signal);
 if (value.plan.id !== id) {throw new Error('Preview reference changed');}
 return checkScope(value, owner, draft);
}

function checkScope(value: TWritePreview, owner: string, draft: TDraft): TWritePreview {
 if (value.plan.owner !== owner || value.plan.draftId !== draft.id || value.plan.draftRevision !== draft.revision || value.plan.targetId !== draft.assetId || value.plan.analysisId !== draft.analysisId) {throw new Error('Preview scope changed');}
 if (value.plan.imageIdentity !== draft.baseline.imageIdentity || value.plan.before.latitude !== draft.baseline.latitude || value.plan.before.longitude !== draft.baseline.longitude || value.plan.intended.latitude !== draft.camera?.latitude || value.plan.intended.longitude !== draft.camera?.longitude) {throw new Error('Preview decision changed');}
 return value;
}
