import {isRecord} from '@/utils/typeGuards';

import {isDescriptionObservation} from './descriptionWriteTypes';
import {isMirrorSelection} from './mirrorTypes';
import {isUUID} from './selectionTypes';

import type {TDescriptionObservation} from './descriptionWriteTypes';
import type {TMirrorSelection} from './mirrorTypes';



export type TDraftPoint = {latitude: number; longitude: number};
export type TDraftDescription = {language: string; status: 'complete' | 'unavailable'; text: string | null; basis: 'scene_only' | 'candidate'; stale: boolean; factsRevision: number; userSupplied: boolean};
export type TDraftBaseline = {description?: TDescriptionObservation | null; status: 'unavailable' | 'reviewed'; latitude: number | null; longitude: number | null; imageIdentity: string; sourceDigest: string; assetId: string; ownerId: string; checksum: string; type: string; observedAt: string};
export type TDraft = {mirror?: TMirrorSelection; headingMethod?: string; headingUncertainty?: number | null; id: string; analysisId: string; assetId: string; revision: number; state: 'draft' | 'staged' | 'rejected'; camera: TDraftPoint | null; fields: ('gps' | 'description')[]; primaryLanguage?: string; descriptionPolicy?: 'preserve' | 'replace' | 'managed_append'; candidateId: string | null; radius: number | null; radiusBasis: string; radiusStale: boolean; factsRevision: number; heading: number | null; headingStale: boolean; headingUserSupplied: boolean; headingFactsRevision: number; descriptions: TDraftDescription[]; originalSourceDigest: string; baseline: TDraftBaseline; updatedAt: string};
const count = (value: unknown): value is number => Number.isSafeInteger(value) && Number(value) > 0;
const numeric = (value: unknown, min: number, max: number): value is number => typeof value === 'number' && Number.isFinite(value) && value >= min && value <= max;
export function isDraftPoint(value: unknown): value is TDraftPoint {return isRecord(value) && numeric(value.latitude, -90, 90) && numeric(value.longitude, -180, 180);}
export function isDraftBaseline(value: unknown): value is TDraftBaseline {
 return isRecord(value) && (value.description == null || isDescriptionObservation(value.description)) && ['unavailable', 'reviewed'].includes(String(value.status)) &&
 (value.latitude === null || numeric(value.latitude, -90, 90)) && (value.longitude === null || numeric(value.longitude, -180, 180)) &&
 ['imageIdentity', 'sourceDigest', 'assetId', 'ownerId', 'checksum', 'type', 'observedAt'].every(key => typeof value[key] === 'string' && value[key].length <= 256);
}
function isDescription(value: unknown): value is TDraftDescription {
 return isRecord(value) && typeof value.language === 'string' && value.language.length <= 100 && ['complete', 'unavailable'].includes(String(value.status)) &&
 (value.text === null || (typeof value.text === 'string' && new TextEncoder().encode(value.text).length <= 16 * 1024)) && ['scene_only', 'candidate'].includes(String(value.basis)) && typeof value.stale === 'boolean' && count(value.factsRevision) && typeof value.userSupplied === 'boolean';
}
export function isDraft(value: unknown): value is TDraft {
 return isRecord(value) && (value.mirror === undefined || isMirrorSelection(value.mirror)) && (value.headingMethod === undefined || (typeof value.headingMethod === 'string' && value.headingMethod.length <= 256)) && (value.headingUncertainty == null || numeric(value.headingUncertainty, 0, 180)) && [value.id, value.analysisId, value.assetId].every(isUUID) && count(value.revision) && count(value.factsRevision) && count(value.headingFactsRevision) &&
 ['draft', 'staged', 'rejected'].includes(String(value.state)) && (value.camera === null || isDraftPoint(value.camera)) &&
 Array.isArray(value.fields) && value.fields.length <= 2 && value.fields.every(field => field === 'gps' || field === 'description') && new Set(value.fields).size === value.fields.length &&
 (value.primaryLanguage === undefined || (typeof value.primaryLanguage === 'string' && value.primaryLanguage.length <= 100)) && (value.descriptionPolicy === undefined || ['preserve', 'replace', 'managed_append'].includes(String(value.descriptionPolicy))) &&
 (value.candidateId === null || (typeof value.candidateId === 'string' && value.candidateId.length <= 128)) &&
 (value.radius === null || numeric(value.radius, 0, Number.MAX_VALUE)) && (value.heading === null || (numeric(value.heading, 0, 360) && value.heading < 360)) &&
 ['radiusStale', 'headingStale', 'headingUserSupplied'].every(key => typeof value[key] === 'boolean') && typeof value.radiusBasis === 'string' &&
 Array.isArray(value.descriptions) && value.descriptions.length <= 20 && value.descriptions.every(isDescription) && new Set(value.descriptions.map(d => d.language)).size === value.descriptions.length &&
 typeof value.originalSourceDigest === 'string' && value.originalSourceDigest.length === 64 && isDraftBaseline(value.baseline) && typeof value.updatedAt === 'string' && Number.isFinite(Date.parse(value.updatedAt));
}
