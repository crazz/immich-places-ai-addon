import {isRecord} from '@/utils/typeGuards';

import {isUUID} from './selectionTypes';
import {isPreviewGPS} from './writePreviewTypes';

import type {TPreviewGPS} from './writePreviewTypes';

export type TStackReview = {id: string; owner: string; installation: string; draftId: string; draftRevision: number; analyzedId: string; imageIdentity: string; stackId: string; expiresAt: string; candidates: string[]; targets: {assetId: string; imageIdentity: string; before: TPreviewGPS}[]};
export function isStackReview(value: unknown): value is TStackReview {
 if (!isRecord(value) || !Array.isArray(value.targets) || !Array.isArray(value.candidates)) {return false;}
 const candidates = value.candidates;
 const ids = value.targets.map(target => isRecord(target) ? target.assetId : null);
 if (new Set(ids).size !== ids.length || !ids.includes(value.analyzedId) || new Set(value.candidates).size !== value.candidates.length || !value.candidates.includes(value.analyzedId) || !ids.every(id => candidates.includes(id))) {return false;}
 if (!value.targets.every((target, index) => isRecord(target) && (index === 0 || String(ids[index - 1]) < String(target.assetId)) && (target.assetId !== value.analyzedId || target.imageIdentity === value.imageIdentity))) {return false;}
 return isRecord(value) && [value.id, value.installation, value.draftId, value.analyzedId].every(isUUID) && (value.stackId === '' || isUUID(value.stackId)) && typeof value.owner === 'string' && typeof value.imageIdentity === 'string' && /^v1:[a-f0-9]{64}$/.test(value.imageIdentity) && Number.isSafeInteger(value.draftRevision) && Number(value.draftRevision) > 0 && typeof value.expiresAt === 'string' && Number.isFinite(Date.parse(value.expiresAt)) && Array.isArray(value.candidates) && value.candidates.every(isUUID) && Array.isArray(value.targets) && value.targets.length > 0 && value.targets.length <= 50 && value.targets.every(target => isRecord(target) && isUUID(target.assetId) && typeof target.imageIdentity === 'string' && /^v1:[a-f0-9]{64}$/.test(target.imageIdentity) && isPreviewGPS(target.before));
}
