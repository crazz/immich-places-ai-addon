import {savedDraft} from './draft';

import type {TDraft} from '../draftTypes';
import type {TWritePreview} from '../writePreviewTypes';

export function stagedPreviewDraft(): TDraft {
 const draft = savedDraft();
 return {...draft, state: 'staged', camera: {latitude: 0, longitude: 12}, fields: ['gps'], radius: 5000, headingStale: true, baseline: {...draft.baseline, status: 'reviewed', latitude: 0, imageIdentity: `v1:${'a'.repeat(64)}`}};
}

export function savedPreview(): TWritePreview {
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
