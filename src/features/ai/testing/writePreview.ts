import {savedDraft} from './draft';

import type {TDraft} from '../draftTypes';
import type {TStackReview} from '../stackReviewTypes';
import type {TTargetOutcome, TWriteOperation} from '../writeOperationTypes';
import type {TGPSWritePlan, TStackWritePlan, TStandardWritePlan, TWritePreview} from '../writePreviewTypes';

export function stagedPreviewDraft(): TDraft {
 const draft = savedDraft();
 return {...draft, state: 'staged', camera: {latitude: 0, longitude: 12}, fields: ['gps'], radius: 5000, headingStale: true, baseline: {...draft.baseline, status: 'reviewed', latitude: 0, imageIdentity: `v1:${'a'.repeat(64)}`}};
}

export function savedPreview(): TWritePreview & {plan: TGPSWritePlan} {
 const draft = stagedPreviewDraft();
 const now = Date.now();
 return {
status: 'usable',
diff: 'changed',
digest: 'b'.repeat(64),
plan: {
  version: 'gps-preview-v1',
comparisonPolicy: 'exact-nullable-gps-v1',
id: '88888888-8888-4888-8888-888888888888',
owner: 'owner',
installation: '77777777-7777-4777-8777-777777777777',
  draftId: draft.id,
draftRevision: draft.revision,
analysisId: draft.analysisId,
imageIdentity: draft.baseline.imageIdentity,
targetId: draft.assetId,
fields: ['gps'],
before: {latitude: 0, longitude: null},
intended: draft.camera!,
observedAt: new Date(now).toISOString(),
createdAt: new Date(now).toISOString(),
expiresAt: new Date(now + 300_000).toISOString()
 }
};
}

export function descriptionPreview(draft: TDraft): TWritePreview & {plan: TStandardWritePlan} {
 const original = savedPreview();
 const plan: Omit<TGPSWritePlan, 'before' | 'intended'> & Partial<Pick<TGPSWritePlan, 'before' | 'intended'>> = {...original.plan};
 delete plan.before; delete plan.intended;
 return {...original, plan: {...plan, version: 'standard-preview-v2', comparisonPolicy: 'selected-standard-fields-v2', draftRevision: draft.revision, fields: ['description'], policyId: 'c'.repeat(64), description: {before: draft.baseline.description!, intended: draft.descriptions[0].text!, language: 'uk', policy: 'replace'}}};
}

export const stackIDs = ['44444444-4444-4444-8444-444444444444', '55555555-5555-4555-8555-555555555555', '66666666-6666-4666-8666-666666666666', '99999999-9999-4999-8999-999999999999'];
export function stackReview(ids = [savedPreview().plan.targetId]): TStackReview {
 const {plan} = savedPreview();
 return {id: '11111111-1111-4111-8111-111111111111', owner: plan.owner, installation: plan.installation, draftId: plan.draftId, draftRevision: plan.draftRevision, analyzedId: plan.targetId, imageIdentity: plan.imageIdentity, stackId: '22222222-2222-4222-8222-222222222222', expiresAt: plan.expiresAt, candidates: [plan.targetId, ...stackIDs].sort(), targets: [...ids].sort().map(assetId => ({assetId, imageIdentity: plan.imageIdentity, before: assetId === plan.targetId ? plan.before : {latitude: null, longitude: 7}}))};
}
export function stackPreview(ids = [savedPreview().plan.targetId, stackIDs[0]]): TWritePreview & {plan: TStackWritePlan} {
 const preview = savedPreview(); const review = stackReview(ids);
 return {...preview, plan: {...preview.plan, version: 'stack-preview-v3' as const, comparisonPolicy: 'selected-stack-targets-v3' as const, policyId: 'c'.repeat(64), manifest: {stackId: review.stackId, reviewId: review.id, targets: review.targets.map(target => ({...target, fields: ['gps' as const], intended: preview.plan.intended}))}}};
}

export function stackOperation(): TWriteOperation & {plan: TStackWritePlan; targets: TTargetOutcome[]} {
 const preview = stackPreview([savedPreview().plan.targetId, ...stackIDs]);
 return {id: '99999999-9999-4999-8999-999999999999', plan: preview.plan, digest: preview.digest, status: 'partial', code: 'STACK_PARTIAL', approvedAt: new Date().toISOString(), attempts: 0, generation: 0, observed: null, verified: false, refreshed: false, noop: false, settled: false, events: [], targets: preview.plan.manifest.targets.map((target, index) => ({assetId: target.assetId, status: ['succeeded', 'conflict', 'failed', 'expired', 'verifying'][index], code: ['GPS_VERIFIED', 'IMMICH_CONFLICT', 'WRITE_FAILED', 'PREVIEW_EXPIRED', 'RECONCILIATION_REQUIRED'][index], attempts: index === 3 ? 0 : 1, generation: index === 3 ? 0 : 1, observed: index === 0 ? target.intended : null, verified: index === 0, refreshed: index === 0, noop: index === 0, settled: index !== 4, fields: [{field: 'gps', status: index === 0 ? 'verified' : index === 1 ? 'conflict' : 'unavailable', wasVerified: index === 0, ...(index === 0 ? {gps: target.intended} : {})}], events: []}))};
}
