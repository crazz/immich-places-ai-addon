import {act, render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import {StandardWriteOutcome} from './StandardWriteOutcome';
import {descriptionDraft} from './testing/draft';
import {descriptionPreview} from './testing/writePreview';
import {WriteOperation} from './WriteOperation';

import type {TWriteOperation} from './writeOperationTypes';

afterEach(() => {vi.unstubAllGlobals(); sessionStorage.clear();});

it('C20 reloads partial field outcomes without marking a newer draft saved or discarding uncertain authority', async () => {
 const user = userEvent.setup(); const original = {...descriptionDraft(), state: 'staged' as const, fields: ['gps' as const, 'description' as const], primaryLanguage: 'uk', descriptionPolicy: 'replace' as const, camera: {latitude: 0, longitude: 12}};
 const preview = descriptionPreview(original); preview.plan.fields = ['gps', 'description']; preview.plan.before = {latitude: null, longitude: null}; preview.plan.intended = original.camera;
 const operation: TWriteOperation = {id: '99999999-9999-4999-8999-999999999999', plan: preview.plan, digest: preview.digest, status: 'partial', code: 'STANDARD_PARTIAL', approvedAt: new Date().toISOString(), attempts: 1, generation: 1, observed: original.camera, verified: false, refreshed: true, noop: false, settled: false, events: [], fields: [{field: 'gps', status: 'verified', wasVerified: true, gps: original.camera}, {field: 'description', status: 'baseline', wasVerified: false, description: original.baseline.description!}]};
 const draft = {...original, revision: 2, state: 'draft' as const}; const refresh = vi.fn();
 const pending = {previewId: '44444444-4444-4444-8444-444444444444', digest: 'e'.repeat(64), idempotencyKey: 'f'.repeat(32)};
 const key = `ai-write-confirmation:owner:${draft.id}`; sessionStorage.setItem(key, JSON.stringify(pending));
 let reply: (value: Response) => void = () => undefined;
 const request = vi.fn<(path: string, init: RequestInit) => Promise<Response>>(async (path: string) => {
  if (path.includes('?draftId=')) {return new Response(JSON.stringify({items: [{id: operation.id, status: operation.status, draftRevision: 1, approvedAt: operation.approvedAt}], nextCursor: ''}));}
  if (path.includes('/by-key/')) {return new Promise<Response>(resolve => {reply = resolve;});}
  return new Response(JSON.stringify(operation));
 }); vi.stubGlobal('fetch', request);
 render(<WriteOperation owner={'owner'} draft={draft} preview={null} onVerified={refresh} />);
 await user.selectOptions(await screen.findByRole('combobox', {name: 'Saved selected fields operations'}), operation.id);
 expect(await screen.findByText('Selected fields operation: partial')).toBeVisible();
 expect(screen.getByText('GPS: verified')).toBeVisible(); expect(screen.getByText('Description: baseline')).toBeVisible();
 expect(screen.getByText('GPS verified and local catalog updated; selected fields remain incomplete.')).toBeVisible();
 expect(screen.getByText('This operation belongs to revision 1; current draft revision 2 is not marked saved.')).toBeVisible();
 expect(screen.getByText(/Prior request completion is unknown/)).toBeVisible();
 await waitFor(() => expect(refresh).toHaveBeenCalledTimes(1));
 await act(async () => reply(new Response(JSON.stringify(operation))));
 expect(sessionStorage.getItem(key)).toBe(JSON.stringify(pending));
 expect(screen.getByRole('button', {name: 'Reconcile confirmation'})).toBeVisible();
 expect(screen.queryByRole('button', {name: 'Retry selected fields write'})).not.toBeInTheDocument();
 expect(request.mock.calls.every(([, init]) => init.method === 'GET')).toBe(true);
});

it('retains previously verified evidence after a later field conflict without reporting current success', async () => {
 const draft = {...descriptionDraft(), state: 'staged' as const, fields: ['description' as const], primaryLanguage: 'uk', descriptionPolicy: 'replace' as const}; const preview = descriptionPreview(draft);
 const op: TWriteOperation = {id: draft.id, plan: preview.plan, digest: preview.digest, status: 'conflict', code: 'DESCRIPTION_CONFLICT', approvedAt: new Date().toISOString(), attempts: 1, generation: 1, observed: null, verified: false, refreshed: false, noop: false, settled: true, events: [], fields: [{field: 'description', status: 'conflict', wasVerified: true, description: {presence: 'null', value: ''}}]};
 vi.stubGlobal('fetch', vi.fn(async (path: string) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [{id: op.id, status: op.status, draftRevision: 1, approvedAt: op.approvedAt}], nextCursor: ''} : op))));
 render(<WriteOperation owner={'owner'} draft={draft} preview={null} />);
 expect(await screen.findByText('Description: conflict')).toBeVisible();
 expect(screen.getByText('Description was previously verified; this is retained evidence, not current verification.')).toBeVisible();
 expect(screen.getByText('Observed description presence: null')).toBeVisible();
 expect(screen.getByLabelText('Observed description').textContent).toBe('');
 expect(screen.getByText('Selected fields are not fully verified saved.')).toBeVisible();
 expect(screen.queryByText('Description verified; no GPS catalog refresh is needed.')).not.toBeInTheDocument();
 expect(screen.queryByRole('button', {name: 'Retry selected fields write'})).not.toBeInTheDocument();
});

it('distinguishes all-field verification from pending GPS catalog refresh and displays selected GPS observations', () => {
 const draft = descriptionDraft(); const preview = descriptionPreview(draft); preview.plan.fields = ['gps', 'description']; preview.plan.before = {latitude: null, longitude: null}; preview.plan.intended = {latitude: 0, longitude: 12};
 const operation: TWriteOperation = {id: draft.id, plan: preview.plan, digest: preview.digest, status: 'verifying', code: 'LOCAL_REFRESH_PENDING', approvedAt: new Date().toISOString(), attempts: 1, generation: 1, observed: preview.plan.intended, verified: true, refreshed: false, noop: false, settled: true, events: [], fields: [{field: 'gps', status: 'verified', wasVerified: true, gps: preview.plan.intended}, {field: 'description', status: 'verified', wasVerified: true, description: {presence: 'value', value: preview.plan.description.intended}}]};
 const view = render(<StandardWriteOutcome operation={operation} />);
 expect(screen.getByText('Observed GPS: 0, 12')).toBeVisible();
 expect(screen.getByText('Selected fields verified upstream; GPS catalog refresh pending.')).toBeVisible();
 expect(screen.getByLabelText('Observed description').textContent).toBe(preview.plan.description.intended);
 view.rerender(<StandardWriteOutcome operation={{...operation, status: 'succeeded', refreshed: true}} />);
 expect(screen.getByText('Selected fields verified and GPS catalog updated.')).toBeVisible();
 expect(screen.queryByText(/refresh pending/)).not.toBeInTheDocument();
});

it('uses selected-field wording for a definitive description write rejection and releases only that identity', async () => {
 const user = userEvent.setup(); const draft = {...descriptionDraft(), state: 'staged' as const, fields: ['description' as const], primaryLanguage: 'uk', descriptionPolicy: 'replace' as const}; const preview = descriptionPreview(draft);
 const request = vi.fn(async (_path: string, init: RequestInit) => new Response(JSON.stringify(init.method === 'POST' ? {code: 'WRITE_DISABLED'} : {items: [], nextCursor: ''}), {status: init.method === 'POST' ? 503 : 200})); vi.stubGlobal('fetch', request);
 render(<WriteOperation owner={'owner'} draft={draft} preview={preview} />);
 await user.click(screen.getByRole('button', {name: 'Confirm selected fields write'}));
 expect(await screen.findByText('Selected fields writing is disabled. Saved review and status remain available.')).toBeVisible();
 expect(screen.queryByRole('button', {name: 'Reconcile confirmation'})).not.toBeInTheDocument();
 expect(sessionStorage.length).toBe(0);
 expect(request.mock.calls.filter(([, init]) => init.method === 'POST')).toHaveLength(1);
});
