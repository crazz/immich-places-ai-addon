import {isRecord} from '@/utils/typeGuards';

import {isDraftPoint} from './draftTypes';
import {isUUID} from './selectionTypes';

import type {TDraftPoint} from './draftTypes';

export type TPreviewGPS = {latitude: number | null; longitude: number | null};
export type TPreviewConflict = {before: TPreviewGPS; current: TPreviewGPS; proposed: TDraftPoint};
export type TWritePlan = {version: 'gps-preview-v1'; comparisonPolicy: 'exact-nullable-gps-v1'; id: string; owner: string; installation: string; draftId: string; draftRevision: number; analysisId: string; imageIdentity: string; targetId: string; fields: ['gps']; before: TPreviewGPS; intended: TDraftPoint; observedAt: string; createdAt: string; expiresAt: string};
export type TWritePreview = {plan: TWritePlan; digest: string; status: 'usable' | 'expired' | 'stale'; diff: 'changed' | 'unchanged'};

export function isPreviewGPS(value: unknown): value is TPreviewGPS {
 if (!isRecord(value)) {return false;}
 return (value.latitude === null || (typeof value.latitude === 'number' && Number.isFinite(value.latitude) && Math.abs(value.latitude) <= 90)) &&
 (value.longitude === null || (typeof value.longitude === 'number' && Number.isFinite(value.longitude) && Math.abs(value.longitude) <= 180));
}

export function isPreviewConflict(value: unknown): value is TPreviewConflict {return isRecord(value) && isPreviewGPS(value.before) && isPreviewGPS(value.current) && isDraftPoint(value.proposed);}

export function isWritePreview(value: unknown): value is TWritePreview {
 if (!isRecord(value) || !isRecord(value.plan)) {return false;}
 const plan = value.plan;
 return ['usable', 'expired', 'stale'].includes(String(value.status)) && ['changed', 'unchanged'].includes(String(value.diff)) && typeof value.digest === 'string' && /^[a-f0-9]{64}$/.test(value.digest) &&
 plan.version === 'gps-preview-v1' && plan.comparisonPolicy === 'exact-nullable-gps-v1' && [plan.id, plan.installation, plan.draftId, plan.analysisId, plan.targetId].every(isUUID) &&
 typeof plan.owner === 'string' && plan.owner.length > 0 && plan.owner.length <= 256 && Number.isSafeInteger(plan.draftRevision) && Number(plan.draftRevision) > 0 &&
 typeof plan.imageIdentity === 'string' && /^v1:[a-f0-9]{64}$/.test(plan.imageIdentity) && Array.isArray(plan.fields) && plan.fields.length === 1 && plan.fields[0] === 'gps' && isPreviewGPS(plan.before) && isDraftPoint(plan.intended) &&
 ['observedAt', 'createdAt', 'expiresAt'].every(key => typeof plan[key] === 'string' && Number.isFinite(Date.parse(plan[key]))) && Date.parse(String(plan.expiresAt)) - Date.parse(String(plan.createdAt)) === 300_000;
}
