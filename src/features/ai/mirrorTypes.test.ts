import {expect, it} from 'vitest';

import {isDraft} from './draftTypes';
import {descriptionDraft} from './testing/draft';

it('validates bounded canonical mirror selections and optional direction evidence in drafts', () => {
 const draft = descriptionDraft(); const mirror = {direction: false, precision: false, place: false, languages: ['uk'], provenance: false, model: false};
 expect(isDraft({...draft, mirror})).toBe(true); expect(isDraft({...draft, mirror: {...mirror, direction: true, languages: null}, headingMethod: 'unknown', headingUncertainty: null})).toBe(true);
 for (const invalid of [{...mirror, languages: ['uk', 'uk']}, {...mirror, languages: ['EN-us']}, {...mirror, languages: ['en', 'uk', 'fr', 'de', 'it', 'es', 'pt', 'ja', 'zh']}, {...mirror, languages: []}, {...mirror, model: true}, {...mirror, targetId: draft.assetId}, {...mirror, direction: 'yes'}]) {expect(isDraft({...draft, mirror: invalid})).toBe(false);}
 expect(isDraft({...draft, headingUncertainty: -1})).toBe(false); expect(isDraft({...draft, headingMethod: {}})).toBe(false);
});

it('M05 rejects oversized unsupported and unselected mirror exports while retaining v1 through v3', async () => {
 const {isWritePreview} = await import('./writePreviewTypes'); const {mirrorPreview} = await import('./testing/mirror'); const {savedPreview, descriptionPreview, stackPreview} = await import('./testing/writePreview');
 const preview = mirrorPreview(); expect(isWritePreview(preview)).toBe(true);
 const mirror = preview.plan.mirror;
 for (const change of [
  {...mirror, privateHint: 'hidden'}, {...mirror, before: {present: false, value: {}}}, {...mirror, before: {present: true, value: null}},
  {...mirror, value: {...mirror.value, privateHint: 'hidden'}}, {...mirror, value: {...mirror.value, direction: {heading: 360, method: 'model', uncertainty: null}}},
  {...mirror, value: {...mirror.value, direction: {heading: null, method: 'unknown', uncertainty: 0}}}, {...mirror, value: {...mirror.value, descriptions: {uk: '河'.repeat(23000)}}},
  {...mirror, value: {...mirror.value, descriptions: {uk: 'Current', en: 'Unselected'}}}, {...mirror, value: {...mirror.value, precision: {radius: 5, basis: 'unselected'}}},
  {...mirror, value: {...mirror.value, review: {draftRevision: 1, factsRevision: 0}}}, {...mirror, value: {...mirror.value, recordId: 'foreign'}},
 ]) {expect(isWritePreview({...preview, plan: {...preview.plan, mirror: change}})).toBe(false);}
 for (const legacy of [savedPreview(), descriptionPreview(descriptionDraft()), stackPreview()]) {
  expect(isWritePreview(legacy)).toBe(true); expect(isWritePreview({...legacy, plan: {...legacy.plan, mirror}})).toBe(false);
 }
});

it('validates exact v4 single and stack manifests without fabricated description-only GPS authority', async () => {
 const {isWritePreview} = await import('./writePreviewTypes'); const {mirrorPreview} = await import('./testing/mirror'); const {stackPreview, descriptionPreview} = await import('./testing/writePreview');
 const single = mirrorPreview(); const stacked = {...single, plan: {...stackPreview().plan, version: 'mirror-preview-v4', comparisonPolicy: 'standard-then-metadata-v4', mirror: single.plan.mirror}};
 const description = descriptionPreview(descriptionDraft()); const described = {...single, plan: {...description.plan, version: 'mirror-preview-v4', comparisonPolicy: 'standard-then-metadata-v4', mirror: single.plan.mirror, manifest: {...single.plan.manifest, targets: [{...single.plan.manifest.targets[0], fields: ['description'], before: {latitude: null, longitude: null}, intended: {latitude: 0, longitude: 0}, description: description.plan.description}]}}};
 expect(isWritePreview(stacked)).toBe(true); expect(isWritePreview(described)).toBe(true);
 for (const plan of [
  {...single.plan, policyId: ''}, {...single.plan, manifest: {...single.plan.manifest, extra: true}},
  {...single.plan, manifest: {...single.plan.manifest, targets: [{...single.plan.manifest.targets[0], mirror: single.plan.mirror}]}},
  {...stacked.plan, manifest: {...stacked.plan.manifest, targets: stacked.plan.manifest.targets.map((target, index) => index ? {...target, fields: ['gps', 'description'], description: description.plan.description} : target)}},
  {...described.plan, manifest: {...described.plan.manifest, targets: [{...described.plan.manifest.targets[0], before: 'unavailable'}]}},
 ]) {expect(isWritePreview({...single, plan})).toBe(false);}
});

it('rejects substituted metadata step identity and malformed audit or verification evidence', async () => {
 const {isWriteOperation} = await import('./writeOperationTypes'); const {mirrorOperation} = await import('./testing/mirror'); const {stackOperation} = await import('./testing/writePreview');
 const op = mirrorOperation(); expect(isWriteOperation(op)).toBe(true);
 for (const mirror of [
  {...op.mirror, assetId: '44444444-4444-4444-8444-444444444444'}, {...op.mirror, step: 'standard'}, {...op.mirror, attempts: 3}, {...op.mirror, generation: -1}, {...op.mirror, targets: []},
  {...op.mirror, events: [{code: 'PRIVATE', at: op.approvedAt, attempt: 1, generation: -1}]}, {...op.mirror, events: [{code: 'PRIVATE', at: 'bad', attempt: 1, generation: 3}]},
  {...op.mirror, status: 'succeeded', verified: true, observed: null}, {...op.mirror, status: 'succeeded', verified: true, observed: {present: true, value: {substituted: 'not approved'}}},
 ]) {expect(isWriteOperation({...op, mirror})).toBe(false);}
 expect(isWriteOperation({...op, mirror: undefined})).toBe(false); expect(isWriteOperation({...stackOperation(), mirror: op.mirror})).toBe(false);
 const reordered = {...op.plan.mirror.value, descriptions: {uk: op.plan.mirror.value.descriptions!.uk}};
 expect(isWriteOperation({...op, mirror: {...op.mirror, status: 'verifying', verified: true, settled: false, observed: {present: true, value: reordered}}})).toBe(true);
});

it('retains prior mirror verification when a later current observation conflicts or becomes absent', async () => {
 const {isWriteOperation} = await import('./writeOperationTypes'); const {mirrorOperation} = await import('./testing/mirror'); const op = mirrorOperation();
 for (const observed of [{present: false}, {present: true, value: {external: 'changed'}}]) {
  expect(isWriteOperation({...op, mirror: {...op.mirror, status: 'conflict', code: 'METADATA_CONFLICT', verified: true, observed}})).toBe(true);
  expect(isWriteOperation({...op, mirror: {...op.mirror, status: 'succeeded', code: 'METADATA_VERIFIED', verified: true, observed}})).toBe(false);
  expect(isWriteOperation({...op, mirror: {...op.mirror, status: 'verifying', code: 'METADATA_OBSERVED_UNRESOLVED', verified: true, observed}})).toBe(false);
 }
});

it('preserves bounded raw UTF-8 text and complete selected precision place and provenance exports', async () => {
 const {isWritePreview} = await import('./writePreviewTypes'); const {mirrorPreview} = await import('./testing/mirror'); const preview = mirrorPreview();
 const text = '河\n"\\'.repeat(2000); const selection = {...preview.plan.mirror.selection, precision: true, place: true, provenance: true, model: true};
 const value = {...preview.plan.mirror.value, descriptions: {uk: text}, precision: {radius: 0, basis: 'source_reported'}, place: '  Київ\n河  ', provenance: {mode: 'visual', model: 'model-v1', reviewedAt: preview.plan.createdAt, fields: {direction: 'user', precision: 'model', place: 'model', ['description:uk']: 'user'}}};
 expect(new TextEncoder().encode(text).length).toBe(12000);
 expect(isWritePreview({...preview, plan: {...preview.plan, mirror: {...preview.plan.mirror, selection, value}}})).toBe(true);
 expect(isWritePreview({...preview, plan: {...preview.plan, mirror: {...preview.plan.mirror, selection, value: {...value, descriptions: {uk: '河'.repeat(5462)}}}}})).toBe(false);
});
