import {isRecord} from '@/utils/typeGuards';

import {isDescriptionPlan} from './descriptionWriteTypes';
import {isDraftPoint} from './draftTypes';
import {isMirrorPlan} from './mirrorTypes';
import {isUUID} from './selectionTypes';


import type {TDescriptionPlan} from './descriptionWriteTypes';
import type {TDraftPoint} from './draftTypes';
import type {TMirrorPlan} from './mirrorTypes';

export type TPreviewGPS = {latitude: number | null; longitude: number | null};
export type TPreviewConflict = {before: TPreviewGPS; current: TPreviewGPS; proposed: TDraftPoint};
type TPlanIdentity = { id: string; owner: string; installation: string; draftId: string; draftRevision: number; analysisId: string; imageIdentity: string; targetId: string; observedAt: string; createdAt: string; expiresAt: string};
export type TGPSWritePlan = TPlanIdentity & {version: 'gps-preview-v1'; comparisonPolicy: 'exact-nullable-gps-v1'; fields: ['gps']; before: TPreviewGPS; intended: TDraftPoint};
export type TStandardWritePlan = TPlanIdentity & {version: 'standard-preview-v2'; comparisonPolicy: 'selected-standard-fields-v2'; fields: ('gps' | 'description')[]; before?: TPreviewGPS; intended?: TDraftPoint; description: TDescriptionPlan; policyId: string};
export type TStackTarget = {assetId: string; imageIdentity: string; fields: ('gps' | 'description')[]; before: TPreviewGPS; intended: TDraftPoint; description?: TDescriptionPlan};
export type TStackWritePlan = TPlanIdentity & {version: 'stack-preview-v3'; comparisonPolicy: 'selected-stack-targets-v3'; fields: ('gps' | 'description')[]; before: TPreviewGPS; intended: TDraftPoint; description?: TDescriptionPlan; policyId: string; manifest: {stackId: string; reviewId: string; targets: TStackTarget[]}};
export type TMirrorWritePlan = TPlanIdentity & {version: 'mirror-preview-v4'; comparisonPolicy: 'standard-then-metadata-v4'; fields: ('gps' | 'description')[]; before?: TPreviewGPS; intended?: TDraftPoint; description?: TDescriptionPlan; policyId: string; manifest: TStackWritePlan['manifest']; mirror: TMirrorPlan};
export type TWritePlan = TMirrorWritePlan | TGPSWritePlan | TStandardWritePlan | TStackWritePlan;
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
 const keys = ['version', 'comparisonPolicy', 'id', 'owner', 'installation', 'draftId', 'draftRevision', 'analysisId', 'imageIdentity', 'targetId', 'fields', 'before', 'intended', 'observedAt', 'createdAt', 'expiresAt', ...(plan.version !== 'gps-preview-v1' ? ['description', 'policyId'] : []), ...(['stack-preview-v3', 'mirror-preview-v4'].includes(String(plan.version)) ? ['manifest'] : []), ...(plan.version === 'mirror-preview-v4' ? ['mirror'] : [])];
 if (!Object.keys(plan).every(key => keys.includes(key))) {return false;}
 return ['usable', 'expired', 'stale'].includes(String(value.status)) && ['changed', 'unchanged'].includes(String(value.diff)) && typeof value.digest === 'string' && /^[a-f0-9]{64}$/.test(value.digest) &&
 validPlanFields(plan) && [plan.id, plan.installation, plan.draftId, plan.analysisId, plan.targetId].every(isUUID) &&
 typeof plan.owner === 'string' && plan.owner.length > 0 && plan.owner.length <= 256 && Number.isSafeInteger(plan.draftRevision) && Number(plan.draftRevision) > 0 &&
 typeof plan.imageIdentity === 'string' && /^v1:[a-f0-9]{64}$/.test(plan.imageIdentity) &&
 ['observedAt', 'createdAt', 'expiresAt'].every(key => typeof plan[key] === 'string' && Number.isFinite(Date.parse(plan[key]))) && Date.parse(String(plan.expiresAt)) - Date.parse(String(plan.createdAt)) === 300_000;
}

function validPlanFields(plan: Record<string, unknown>): boolean {
 if (plan.version === 'mirror-preview-v4') {return validMirrorPlan(plan);}
 if (plan.version === 'stack-preview-v3') {return validStackPlan(plan);}
 if (plan.version === 'gps-preview-v1') {return plan.comparisonPolicy === 'exact-nullable-gps-v1' && Array.isArray(plan.fields) && plan.fields.length === 1 && plan.fields[0] === 'gps' && isPreviewGPS(plan.before) && isDraftPoint(plan.intended);}
 return plan.version === 'standard-preview-v2' && plan.comparisonPolicy === 'selected-standard-fields-v2' && Array.isArray(plan.fields) && plan.fields.includes('description') && plan.fields.length <= 2 && new Set(plan.fields).size === plan.fields.length && plan.fields.every(field => field === 'gps' || field === 'description') && (plan.fields.includes('gps') ? isPreviewGPS(plan.before) && isDraftPoint(plan.intended) : plan.before === undefined && plan.intended === undefined) && isDescriptionPlan(plan.description) && typeof plan.policyId === 'string' && /^[a-f0-9]{64}$/.test(plan.policyId);
}

function validStackPlan(plan: Record<string, unknown>): boolean {
 if (plan.comparisonPolicy !== 'selected-stack-targets-v3' || !isRecord(plan.manifest) || !Object.keys(plan.manifest).every(key => ['stackId', 'reviewId', 'targets'].includes(key)) || !isUUID(plan.manifest.stackId) || !isUUID(plan.manifest.reviewId) || !Array.isArray(plan.manifest.targets) || plan.manifest.targets.length < 2 || plan.manifest.targets.length > 50 || !Array.isArray(plan.fields) || plan.fields[0] !== 'gps' || plan.fields.length > 2 || (plan.fields.length === 2 && plan.fields[1] !== 'description') || !isPreviewGPS(plan.before) || !isDraftPoint(plan.intended) || typeof plan.policyId !== 'string' || !/^[a-f0-9]{64}$/.test(plan.policyId)) {return false;}
 if (plan.fields.includes('description') ? !isDescriptionPlan(plan.description) : plan.description !== undefined) {return false;}
 const ids = plan.manifest.targets.map(target => isRecord(target) ? target.assetId : null);
 if (new Set(ids).size !== ids.length || !ids.includes(plan.targetId)) {return false;}
 const before = plan.before; const intended = plan.intended; const fields = plan.fields;
 return plan.manifest.targets.every((target, index) => {
  if (!isRecord(target) || !Object.keys(target).every(key => ['assetId', 'imageIdentity', 'fields', 'before', 'intended', 'description'].includes(key)) || !isUUID(target.assetId) || (index > 0 && String(ids[index - 1]) >= target.assetId) || typeof target.imageIdentity !== 'string' || !/^v1:[a-f0-9]{64}$/.test(target.imageIdentity) || !Array.isArray(target.fields) || !isPreviewGPS(target.before) || !isDraftPoint(target.intended) || target.intended.latitude !== intended.latitude || target.intended.longitude !== intended.longitude) {return false;}
  if (target.assetId !== plan.targetId) {return target.fields.length === 1 && target.fields[0] === 'gps' && target.description === undefined;}
  return target.imageIdentity === plan.imageIdentity && target.before.latitude === before.latitude && target.before.longitude === before.longitude && target.fields.length === fields.length && target.fields.every((field, fieldIndex) => field === fields[fieldIndex]) && JSON.stringify(target.description) === JSON.stringify(plan.description);
 });
}

function validMirrorPlan(plan: Record<string, unknown>): boolean {
 if (typeof plan.policyId !== 'string' || !/^[a-f0-9]{64}$/.test(plan.policyId) || plan.comparisonPolicy !== 'standard-then-metadata-v4' || !isMirrorPlan(plan.mirror) || !isRecord(plan.manifest) || !Array.isArray(plan.manifest.targets) || plan.mirror.value.review.draftRevision !== plan.draftRevision) {return false;}
 if (plan.manifest.targets.length > 1) {return validStackPlan({...plan, comparisonPolicy: 'selected-stack-targets-v3'});}
 if (!Object.keys(plan.manifest).every(key => ['stackId', 'reviewId', 'targets'].includes(key))) {return false;}
 const target = plan.manifest.targets[0];
 const hasDescription = Array.isArray(plan.fields) && plan.fields.includes('description');
 return plan.manifest.targets.length === 1 && plan.manifest.stackId === '' && plan.manifest.reviewId === '' && isRecord(target) && Object.keys(target).every(key => ['assetId', 'imageIdentity', 'fields', 'before', 'intended', 'description'].includes(key)) && isPreviewGPS(target.before) && isDraftPoint(target.intended) && (hasDescription || plan.description === undefined) && target.assetId === plan.targetId && target.imageIdentity === plan.imageIdentity && JSON.stringify(target.fields) === JSON.stringify(plan.fields) && JSON.stringify(target.description) === JSON.stringify(plan.description) && validPlanFields({...plan, version: hasDescription ? 'standard-preview-v2' : 'gps-preview-v1', comparisonPolicy: hasDescription ? 'selected-standard-fields-v2' : 'exact-nullable-gps-v1'}) && (Array.isArray(plan.fields) && !plan.fields.includes('gps') || JSON.stringify(target.before) === JSON.stringify(plan.before) && JSON.stringify(target.intended) === JSON.stringify(plan.intended));
}
