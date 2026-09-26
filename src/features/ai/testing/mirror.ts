import {descriptionDraft} from './draft';
import {savedPreview, stagedPreviewDraft} from './writePreview';

import type {TDraft} from '../draftTypes';
import type {TMirrorOutcome,TMirrorSelection} from '../mirrorTypes';
import type {TTargetOutcome, TWriteOperation} from '../writeOperationTypes';
import type {TMirrorWritePlan,TWritePreview} from '../writePreviewTypes';



export function mirroredDraft(): TDraft & {mirror: TMirrorSelection} {
 return {...stagedPreviewDraft(), headingStale: false, radiusStale: false, descriptions: descriptionDraft().descriptions, mirror: {direction: true, precision: false, place: false, languages: ['uk'], provenance: false, model: false}};
}
export function mirrorPreview(draft = mirroredDraft()): TWritePreview & {plan: TMirrorWritePlan} {
 const preview = savedPreview(); const recordId = '11111111-1111-4111-8111-111111111111';
 return {...preview, plan: {...preview.plan, draftRevision: draft.revision, version: 'mirror-preview-v4' as const, comparisonPolicy: 'standard-then-metadata-v4' as const, policyId: 'c'.repeat(64), manifest: {stackId: '', reviewId: '', targets: [{assetId: draft.assetId, imageIdentity: draft.baseline.imageIdentity, fields: draft.fields, before: preview.plan.before, intended: preview.plan.intended}]}, mirror: {key: 'immich-places-ai-addon' as const, recordId, before: {present: false}, selection: draft.mirror, disclosure: 'asset-readers-v1' as const, value: {schemaVersion: 1 as const, application: 'immich-places-ai-addon' as const, recordId, review: {draftRevision: draft.revision, factsRevision: draft.factsRevision}, direction: {heading: null, method: 'unknown', uncertainty: null}, descriptions: {uk: draft.descriptions[0].text!}}}}};
}

export function mirrorOperation(): TWriteOperation & {plan: TMirrorWritePlan; mirror: TMirrorOutcome; targets: TTargetOutcome[]} {
 const preview = mirrorPreview();
 return {id: '99999999-9999-4999-8999-999999999999', plan: preview.plan, digest: preview.digest, status: 'partial', code: 'MIRROR_PARTIAL', approvedAt: preview.plan.createdAt, attempts: 0, generation: 0, observed: null, verified: false, refreshed: true, noop: false, settled: true, events: [], targets: preview.plan.manifest.targets.map(target => ({assetId: target.assetId, status: 'succeeded', code: 'GPS_VERIFIED', attempts: 1, generation: 1, observed: target.intended, verified: true, refreshed: true, noop: false, settled: true, fields: [{field: 'gps' as const, status: 'verified' as const, wasVerified: true, gps: target.intended}], events: []})), mirror: {assetId: preview.plan.targetId, step: 'metadata' as const, status: 'retryable', code: 'METADATA_FAILED', attempts: 1, generation: 3, observed: {present: false}, verified: false, noop: false, settled: true, events: []}};
}
