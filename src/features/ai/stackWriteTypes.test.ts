import {expect, it} from 'vitest';

import {stackPreview} from './testing/writePreview';
import {isWritePreview} from './writePreviewTypes';

it('S04 rejects repeated or oversized immutable targets without truncation', () => {
 const preview = stackPreview(); const target = preview.plan.manifest.targets[0];
 expect(isWritePreview(preview)).toBe(true);
 for (const targets of [[target, target], Array.from({length: 51}, (_, index) => ({...target, assetId: `00000000-0000-4000-8000-${String(index).padStart(12, '0')}`}))]) {
  expect(isWritePreview({...preview, plan: {...preview.plan, manifest: {...preview.plan.manifest, targets}}})).toBe(false);
 }
});

it('rejects substituted v3 field source and GPS matrices and legacy manifest injection', () => {
 const preview = stackPreview(); const {plan} = preview; const [primary, sibling] = plan.manifest.targets;
 const altered = [
  [sibling, primary],
  [{...primary, imageIdentity: `v1:${'f'.repeat(64)}`}, sibling],
  [{...primary, before: {latitude: null, longitude: null}}, sibling],
  [primary, {...sibling, intended: {latitude: 1, longitude: 12}}],
  [primary, {...sibling, fields: ['gps', 'description'], description: {before: {presence: 'absent', value: ''}, intended: 'sibling', language: 'en', policy: 'replace'}}],
  [primary, {...sibling, fields: ['gps'], heading: 90}],
  [primary, {...sibling, fields: ['gps', 'gps']}]
 ];
 for (const targets of altered) {expect(isWritePreview({...preview, plan: {...plan, manifest: {...plan.manifest, targets}}})).toBe(false);}
 expect(isWritePreview({...preview, plan: {...plan, fields: ['description']}})).toBe(false);
 expect(isWritePreview({...preview, plan: {...plan, version: 'gps-preview-v1', comparisonPolicy: 'exact-nullable-gps-v1'}})).toBe(false);
});

it('binds preview publication to exact reviewed selection and observations', async () => {
 const {vi} = await import('vitest'); const {createWritePreview} = await import('./writePreviewApi');
 const {stagedPreviewDraft} = await import('./testing/writePreview'); const {stackReview} = await import('./testing/writePreview');
 const draft = stagedPreviewDraft(); const preview = stackPreview(); const review = stackReview(preview.plan.manifest.targets.map(target => target.assetId));
 try {
  vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(preview))));
  await expect(createWritePreview('owner', draft)).rejects.toThrow();
  await expect(createWritePreview('owner', draft, undefined, review)).resolves.toEqual(preview);
  for (const changed of [{...review, id: '99999999-9999-4999-8999-999999999999'}, {...review, targets: review.targets.slice(0, 1)}, {...review, targets: review.targets.map(target => ({...target, before: {latitude: 5, longitude: 7}}))}]) {
   await expect(createWritePreview('owner', draft, undefined, changed)).rejects.toThrow();
  }
 } finally {vi.unstubAllGlobals();}
});

it('rejects malformed target outcomes and aggregate authority substitution', async () => {
 const {stackOperation} = await import('./testing/writePreview'); const {isWriteOperation} = await import('./writeOperationTypes');
 const operation = stackOperation();
 expect(isWriteOperation(operation)).toBe(true);
 for (const invalid of [
  {...operation, targets: operation.targets.slice(1)},
  {...operation, targets: [...operation.targets].reverse()},
  {...operation, attempts: 1},
  {...operation, generation: 1},
  {...operation, fields: [{field: 'gps', status: 'pending', wasVerified: false}]},
  {...operation, targets: operation.targets.map((target, index) => index === 0 ? {...target, fields: []} : target)},
  {...operation, targets: operation.targets.map((target, index) => index === 0 ? {...target, fields: [{field: 'gps', status: 'verified', wasVerified: false, gps: target.observed}]} : target)},
  {...operation, targets: operation.targets.map((target, index) => index === 0 ? {...target, attempts: 3} : target)}
 ]) {expect(isWriteOperation(invalid)).toBe(false);}
});

it('binds primary-only stack description text to the reviewed CH20 decision', async () => {
 const {vi} = await import('vitest'); const {createWritePreview} = await import('./writePreviewApi');
 const {descriptionDraft} = await import('./testing/draft'); const {stagedPreviewDraft} = await import('./testing/writePreview'); const {stackReview} = await import('./testing/writePreview');
 const draft = {...stagedPreviewDraft(), descriptions: descriptionDraft().descriptions, fields: ['gps', 'description'] as ('gps' | 'description')[], primaryLanguage: 'uk', descriptionPolicy: 'replace' as const};
 draft.baseline.description = descriptionDraft().baseline.description;
 const preview = stackPreview(); const review = stackReview(preview.plan.manifest.targets.map(target => target.assetId));
 try {
  for (const intended of [draft.descriptions[0].text!, 'substituted text']) {
   const description = {before: draft.baseline.description!, intended, language: 'uk', policy: 'replace'};
   const plan = {...preview.plan, fields: draft.fields, description, manifest: {...preview.plan.manifest, targets: preview.plan.manifest.targets.map(target => target.assetId === draft.assetId ? {...target, fields: draft.fields, description} : target)}};
   vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({...preview, plan}))));
   if (intended === draft.descriptions[0].text) {await expect(createWritePreview('owner', draft, undefined, review)).resolves.toMatchObject({plan});}
   else {await expect(createWritePreview('owner', draft, undefined, review)).rejects.toThrow();}
  }
 } finally {vi.unstubAllGlobals();}
});

it('keeps v3 target authority and expired status out of legacy operation records', async () => {
 const {stackOperation, savedPreview} = await import('./testing/writePreview'); const {isWriteOperation} = await import('./writeOperationTypes');
 const op = stackOperation(); const {targets, ...legacy} = {...op, plan: savedPreview().plan, status: 'succeeded'};
 expect(isWriteOperation(legacy)).toBe(true);
 expect(isWriteOperation({...legacy, targets})).toBe(false);
 expect(isWriteOperation({...legacy, status: 'expired'})).toBe(false);
});
