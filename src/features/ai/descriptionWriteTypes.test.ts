import {afterEach, expect, it, vi} from 'vitest';

import {isDraft} from './draftTypes';
import {descriptionDraft} from './testing/draft';
import {descriptionPreview} from './testing/writePreview';
import {isWriteOperation} from './writeOperationTypes';
import {createWritePreview} from './writePreviewApi';
import {isWritePreview} from './writePreviewTypes';

afterEach(() => vi.unstubAllGlobals());

it('C03 rejects unusable description selections without publishing a plan or changing draft values', async () => {
 const original = {...descriptionDraft(), state: 'staged' as const, fields: ['description' as const], primaryLanguage: 'uk', descriptionPolicy: 'replace' as const};
 for (const change of [{text: ''}, {text: '   \n'}, {status: 'unavailable' as const, text: null}, {stale: true}, {basis: 'candidate' as const, factsRevision: 2}]) {
  const draft = {...original, descriptions: [{...original.descriptions[0], ...change}]}; const unchanged = JSON.stringify(draft);
  vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(descriptionPreview(draft)))));
  await expect(createWritePreview('owner', draft)).rejects.toThrow();
  expect(JSON.stringify(draft)).toBe(unchanged);
 }
 expect(isDraft({...original, primaryLanguage: ['uk', 'en']})).toBe(false);
 const empty = {...original, fields: []};
 vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(descriptionPreview(original)))));
 await expect(createWritePreview('owner', empty)).rejects.toThrow();
});

it('C06 ignores unselected GPS changes but rejects unavailable or malformed text baselines', async () => {
 const original = {...descriptionDraft(), state: 'staged' as const, fields: ['description' as const], primaryLanguage: 'uk', descriptionPolicy: 'replace' as const};
 const preview = descriptionPreview(original);
 const draft = {...original, camera: {latitude: 40, longitude: 14}, baseline: {...original.baseline, latitude: 40, longitude: 14}};
 vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(preview))));
 await expect(createWritePreview('owner', draft)).resolves.toEqual(preview);
 for (const before of [undefined, null, {}, {presence: 'absent', value: 'not empty'}, {presence: 'value', value: 1}]) {
  const changed = {...preview, plan: {...preview.plan, description: {...preview.plan.description, before}}};
  vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(changed))));
  await expect(createWritePreview('owner', draft)).rejects.toThrow();
 }
 vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(preview))));
 await expect(createWritePreview('owner', {...draft, baseline: {...draft.baseline, description: undefined}})).rejects.toThrow();
 expect(draft.baseline.description).toEqual(original.baseline.description);
});

it('rejects unsupported plan members, malformed append blocks and substituted exact description decisions', async () => {
 const draft = {...descriptionDraft(), state: 'staged' as const, fields: ['description' as const], primaryLanguage: 'uk', descriptionPolicy: 'replace' as const}; const preview = descriptionPreview(draft);
 for (const change of [{unexpected: true}, {fields: ['description', 'description']}, {fields: ['heading']}, {before: {latitude: 0, longitude: 0}}, {policyId: 'not-a-hash'}, {description: {...preview.plan.description, unexpected: true}}, {description: {...preview.plan.description, intended: '文'.repeat(6000)}}]) {
  expect(isWritePreview({...preview, plan: {...preview.plan, ...change}})).toBe(false);
 }
 const id = '55555555-5555-4555-8555-555555555555';
 const block = `[[Immich Places AI v1:${id}]]\nLanguage: uk\n${draft.descriptions[0].text}\n[[/Immich Places AI v1:${id}]]`;
 const append = {...preview.plan.description, policy: 'managed_append', lineage: {id, block, hash: 'd'.repeat(64)}, intended: `${preview.plan.description.before.value}\n\n${block}`};
 expect(isWritePreview({...preview, plan: {...preview.plan, description: append}})).toBe(true);
 for (const changedBlock of ['plain text', block.replace('Language: uk', 'Language: en'), block.replace(/\n\[\[\//, '\n[[broken/')]) {
  expect(isWritePreview({...preview, plan: {...preview.plan, description: {...append, intended: changedBlock, lineage: {...append.lineage, block: changedBlock}}}})).toBe(false);
 }
 for (const change of [{intended: draft.descriptions[0].text!.trim()}, {language: 'en'}, {before: {presence: 'value', value: preview.plan.description.before.value.trim()}}]) {
  vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({...preview, plan: {...preview.plan, description: {...preview.plan.description, ...change}}}))));
  await expect(createWritePreview('owner', draft)).rejects.toThrow();
 }
});

it('accepts exact mixed fields and pending outcomes while rejecting substituted field projections', async () => {
 const draft = {...descriptionDraft(), state: 'staged' as const, camera: {latitude: 0, longitude: 12}, fields: ['description' as const, 'gps' as const], primaryLanguage: 'uk', descriptionPolicy: 'replace' as const}; const preview = descriptionPreview(draft);
 preview.plan.fields = ['description', 'gps']; preview.plan.before = {latitude: null, longitude: null}; preview.plan.intended = draft.camera;
 vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(preview))));
 await expect(createWritePreview('owner', draft)).resolves.toEqual(preview);
 const fields = [{field: 'description', status: 'pending', wasVerified: false}, {field: 'gps', status: 'pending', wasVerified: false}];
 const operation = {id: draft.id, plan: preview.plan, digest: preview.digest, status: 'queued', code: '', approvedAt: new Date().toISOString(), attempts: 0, generation: 0, observed: null, verified: false, refreshed: false, noop: false, settled: true, events: [], fields};
 expect(isWriteOperation(operation)).toBe(true);
 for (const changed of [undefined, [fields[0]], [...fields].reverse(), [fields[0], {...fields[1], field: 'heading'}], [fields[0], {...fields[1], status: 'unknown'}], [{...fields[0], gps: draft.camera}, fields[1]], [{...fields[0], description: {presence: 'null', value: 'bad'}}, fields[1]]]) {
  expect(isWriteOperation({...operation, fields: changed})).toBe(false);
 }
});

it('rejects claimed field verification without an observed selected-field value', () => {
 const draft = descriptionDraft(); const preview = descriptionPreview(draft);
 const operation = {id: draft.id, plan: preview.plan, digest: preview.digest, status: 'succeeded', code: 'STANDARD_VERIFIED', approvedAt: new Date().toISOString(), attempts: 1, generation: 1, observed: null, verified: true, refreshed: false, noop: false, settled: true, events: []};
 for (const status of ['verified', 'baseline']) {
  expect(isWriteOperation({...operation, fields: [{field: 'description', status, wasVerified: true}]})).toBe(false);
 }
 expect(isWriteOperation({...operation, verified: false, status: 'conflict', fields: [{field: 'description', status: 'conflict', wasVerified: true}]})).toBe(true);
 expect(isWriteOperation({...operation, fields: [{field: 'description', status: 'verified', wasVerified: true, description: {presence: 'value', value: preview.plan.description.intended}}]})).toBe(true);
});
