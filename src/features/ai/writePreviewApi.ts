import {backendFetch, parseJSON} from '@/shared/services/backendApi.fetch';
import {getBackendBaseURL} from '@/utils/backendUrls';
import {isRecord} from '@/utils/typeGuards';

import {AIRequestError} from './aiRequest';
import {raceAbort, withCompleteOperationDeadline} from './completeOperation';
import {isCurrentDescription} from './descriptionWriteTypes';
import {mirrorMatchesDraft} from './mirrorDraft';
import {isPreviewConflict, isWritePreview} from './writePreviewTypes';

import type {TDraft} from './draftTypes';
import type {TStackReview} from './stackReviewTypes';
import type {TPreviewConflict, TWritePreview} from './writePreviewTypes';

export class PreviewRequestError extends AIRequestError {
 constructor(code: string, readonly conflict?: TPreviewConflict) {super(code, 'Preview unavailable.');}
}

async function previewRequest(path: string, method: 'GET' | 'POST', input: unknown, signal?: AbortSignal, timeoutMs = 15_000): Promise<TWritePreview> {
 return withCompleteOperationDeadline(timeoutMs, signal, async operationSignal => raceAbort(operationSignal, (async () => {
  const response = await backendFetch(`${getBackendBaseURL()}${path}`, {method, cache: 'no-store', headers: new Headers([['Content-Type', 'application/json']]), body: method === 'POST' ? JSON.stringify(input) : undefined}, {signal: operationSignal, timeoutMs});
  if (!response.ok) {
   let failure: unknown; try {failure = await response.json();} catch {failure = null;}
   throw new PreviewRequestError(isRecord(failure) && typeof failure.code === 'string' ? failure.code : 'REQUEST_FAILED', isRecord(failure) && failure.code === 'IMMICH_CONFLICT' && isPreviewConflict(failure.conflict) ? failure.conflict : undefined);
  }
  return parseJSON(response, isWritePreview, 'Invalid preview response');
 })()));
}

export async function createWritePreview(owner: string, draft: TDraft, signal?: AbortSignal, review?: TStackReview, mirrorDisclosure?: 'asset-readers-v1'): Promise<TWritePreview> {
 if (draft.mirror && mirrorDisclosure !== 'asset-readers-v1') {throw new Error('Mirror disclosure required');}
 const value = await previewRequest('/ai/write-previews', 'POST', {draftId: draft.id, draftRevision: draft.revision, ...(review ? {stackReviewId: review.id} : {}), ...(mirrorDisclosure ? {mirrorDisclosure} : {})}, signal, review && review.targets.length > 1 ? 30_000 : 15_000);
 if (value.plan.version === 'stack-preview-v3' || (value.plan.version === 'mirror-preview-v4' && value.plan.manifest.targets.length > 1)) {
  const manifest = value.plan.manifest;
  if (manifest.reviewId !== review?.id || manifest.stackId !== review.stackId || manifest.targets.length !== review.targets.length || !manifest.targets.every((target, index) => {const observed = review.targets[index]; return target.assetId === observed.assetId && target.imageIdentity === observed.imageIdentity && target.before.latitude === observed.before.latitude && target.before.longitude === observed.before.longitude;})) {throw new Error('Preview target review changed');}
 } else if (review && review.targets.length > 1) {throw new Error('Preview targets changed');}
 return checkScope(value, owner, draft);
}

export async function fetchWritePreview(owner: string, draft: TDraft, id: string, signal?: AbortSignal): Promise<TWritePreview> {
 const value = await previewRequest(`/ai/write-previews/${encodeURIComponent(id)}`, 'GET', undefined, signal);
 if (value.plan.id !== id) {throw new Error('Preview reference changed');}
 return checkScope(value, owner, draft);
}

function checkScope(value: TWritePreview, owner: string, draft: TDraft): TWritePreview {
 if (value.plan.owner !== owner || value.plan.draftId !== draft.id || value.plan.draftRevision !== draft.revision || value.plan.targetId !== draft.assetId || value.plan.analysisId !== draft.analysisId) {throw new Error('Preview scope changed');}
 const plan = value.plan;
 if (plan.version === 'mirror-preview-v4' ? !mirrorMatchesDraft(plan.mirror, draft) : !!draft.mirror) {throw new Error('Preview mirror decision changed');}
 if (plan.imageIdentity !== draft.baseline.imageIdentity || plan.fields.length !== draft.fields.length || !plan.fields.every(field => draft.fields.includes(field))) {throw new Error('Preview decision changed');}
 if (plan.fields.includes('gps') && (plan.before?.latitude !== draft.baseline.latitude || plan.before?.longitude !== draft.baseline.longitude || plan.intended?.latitude !== draft.camera?.latitude || plan.intended?.longitude !== draft.camera?.longitude)) {throw new Error('Preview decision changed');}
 if (plan.version !== 'gps-preview-v1' && plan.description) {
  const chosen = draft.descriptions.find(item => item.language === draft.primaryLanguage);
  if (!isCurrentDescription(chosen, draft.factsRevision) || plan.description.language !== draft.primaryLanguage || plan.description.policy !== draft.descriptionPolicy || plan.description.before.presence !== draft.baseline.description?.presence || plan.description.before.value !== draft.baseline.description?.value || (plan.description.policy === 'replace' ? plan.description.intended !== chosen?.text : plan.description.lineage?.block !== `[[Immich Places AI v1:${plan.description.lineage?.id}]]\nLanguage: ${chosen?.language}\n${chosen?.text}\n[[/Immich Places AI v1:${plan.description.lineage?.id}]]`)) {throw new Error('Preview description changed');}
 }
 return value;
}
