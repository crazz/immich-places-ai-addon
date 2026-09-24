import {isRecord} from '@/utils/typeGuards';

import {draftRequest} from './draftApi';
import {isDraft, isDraftBaseline} from './draftTypes';
import {isUUID} from './selectionTypes';

import type {TDraft, TDraftBaseline} from './draftTypes';

export type TDraftObservation = {id: string; draftId: string; revision: number; baseline: TDraftBaseline; expiresAt: string; previewUrl: string; originalSourceMatches: boolean};
export function isDraftObservation(value: unknown): value is TDraftObservation {
 return isRecord(value) && isUUID(value.id) && isUUID(value.draftId) && Number.isSafeInteger(value.revision) && Number(value.revision) > 0 && isDraftBaseline(value.baseline) && value.baseline.status === 'reviewed' &&
 typeof value.expiresAt === 'string' && Number.isFinite(Date.parse(value.expiresAt)) && typeof value.previewUrl === 'string' && /^\/ai\/jobs\/[0-9a-f-]{36}\/items\/[0-9a-f-]{36}\/thumbnail$/.test(value.previewUrl) && typeof value.originalSourceMatches === 'boolean';
}
export async function observeDraft(draft: TDraft, signal?: AbortSignal): Promise<TDraftObservation> {return draftRequest(`/ai/drafts/${draft.id}/baseline`, 'POST', {}, isDraftObservation, signal, draft.revision);}
export async function acknowledgeDraft(draft: TDraft, observation: TDraftObservation, signal?: AbortSignal): Promise<TDraft> {return draftRequest(`/ai/drafts/${draft.id}/baseline`, 'POST', {observationId: observation.id}, isDraft, signal, draft.revision);}
