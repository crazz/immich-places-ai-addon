import {render, screen} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {StackWriteOutcome} from './StackWriteOutcome';
import {descriptionDraft} from './testing/draft';
import {mirrorOperation, mirroredDraft} from './testing/mirror';
import {descriptionPreview} from './testing/writePreview';

import type {TDraft} from './draftTypes';
import type {TWriteOperation} from './writeOperationTypes';

function descriptionMirror(): {draft: TDraft; operation: TWriteOperation} {
 const draft = {...mirroredDraft(), ...descriptionDraft(), state: 'staged' as const, fields: ['description' as const], primaryLanguage: 'uk', descriptionPolicy: 'replace' as const};
 const original = mirrorOperation(); const standard = descriptionPreview(draft);
 const plan = {...standard.plan, version: 'mirror-preview-v4' as const, comparisonPolicy: 'standard-then-metadata-v4' as const, mirror: original.plan.mirror, manifest: {...original.plan.manifest, targets: [{...original.plan.manifest.targets[0], fields: draft.fields, before: {latitude: null, longitude: null}, intended: {latitude: 0, longitude: 0}, description: standard.plan.description}]}};
 const operation: TWriteOperation = {...original, plan, targets: [{...original.targets[0], code: 'STANDARD_UNCHANGED', refreshed: false, noop: true, observed: null, fields: [{field: 'description', status: 'verified', wasVerified: true, description: {presence: 'value', value: standard.plan.description.intended}}]}]};
 return {draft, operation};
}

afterEach(() => {vi.unstubAllGlobals(); sessionStorage.clear();});

it('describes a text-only standard no-op without claiming GPS verification', () => {
 const {operation} = descriptionMirror();
 render(<StackWriteOutcome operation={operation} />);
 expect(screen.getByText('Description verified unchanged; no write sent for this target.')).toBeVisible();
 expect(screen.queryByText('GPS verified unchanged; no write sent for this target.')).not.toBeInTheDocument();
});

it('retries only the approved description and replays its retained generation with manual GPS pending', async () => {
 const {WriteOperation} = await import('./WriteOperation'); const {default: userEvent} = await import('@testing-library/user-event');
 const user = userEvent.setup(); const {draft, operation} = descriptionMirror();
 operation.status = 'partial'; operation.targets![0] = {...operation.targets![0], status: 'retryable', code: 'WRITE_FAILED', verified: false, noop: false, fields: [{field: 'description', status: 'baseline', wasVerified: false, description: draft.baseline.description ?? undefined}]};
 operation.mirror = {...operation.mirror!, status: 'blocked', code: '', attempts: 0, generation: 0};
 const pending = [draft.assetId]; let posts = 0;
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (init.method === 'POST' && ++posts === 1) {throw new Error('lost retry acknowledgement');}
  return new Response(JSON.stringify(path.includes('?draftId=') ? {items: [{id: operation.id, status: operation.status, draftRevision: draft.revision, approvedAt: operation.approvedAt}], nextCursor: ''} : operation));
 }); vi.stubGlobal('fetch', request);
 const view = render(<WriteOperation owner={'owner'} draft={draft} preview={null} manualPendingIDs={pending} />);
 const retry = await screen.findByRole('button', {name: `Retry description for ${draft.assetId}`});
 expect(retry).toBeEnabled(); await user.click(retry);
 expect(await screen.findByText(`Retry acknowledgement unavailable for ${draft.assetId}. Reconcile status or repeat the same target generation.`)).toBeVisible();
 view.unmount(); render(<WriteOperation owner={'owner'} draft={draft} preview={null} manualPendingIDs={pending} disabled />);
 const replay = await screen.findByRole('button', {name: `Repeat retry generation 1 for ${draft.assetId}`});
 expect(replay).toBeEnabled(); await user.click(replay);
 expect(request.mock.calls.filter(([, init]) => init.method === 'POST').map(([path, init]) => [path.split('/ai')[1], JSON.parse(String(init.body))])).toEqual(Array(2).fill([`/write-operations/${operation.id}/targets/${draft.assetId}/retry`, {generation: 1}]));
 expect(pending).toEqual([draft.assetId]);
 expect(screen.queryByText('Resolve this target’s pending manual GPS before any retry. Read-only status remains available.')).not.toBeInTheDocument();
});
