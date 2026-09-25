import {act, render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import {savedPreview, stagedPreviewDraft} from './testing/writePreview';
import {WriteOperation} from './WriteOperation';

afterEach(() => {vi.unstubAllGlobals(); sessionStorage.clear();});

it('retains an unresolved confirmation when initial saved history arrives late', async () => {
 const draft = stagedPreviewDraft(); const preview = savedPreview(); const user = userEvent.setup();
 const key = `ai-write-confirmation:owner:${draft.id}`;
 const old = {id: '99999999-9999-4999-8999-999999999999', plan: {...preview.plan, id: '66666666-6666-4666-8666-666666666666'}, digest: preview.digest, status: 'canceled', code: 'DRAFT_CHANGED', approvedAt: new Date().toISOString(), attempts: 0, generation: 0, observed: null, verified: false, refreshed: false, noop: false, settled: true, events: []};
 let release: (value: Response) => void = () => undefined;
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (init.method === 'POST') {throw new Error('lost acknowledgement');}
  if (path.includes('?draftId=')) {return new Promise<Response>(resolve => {release = resolve;});}
  return new Response(JSON.stringify(old));
 }); vi.stubGlobal('fetch', request);
 render(<WriteOperation owner={'owner'} draft={draft} preview={preview} />);
 await user.click(screen.getByRole('button', {name: 'Confirm GPS write'}));
 expect(await screen.findByText(/Confirmation acknowledgement unavailable/)).toBeVisible();
 const pending = sessionStorage.getItem(key); expect(pending).not.toBeNull();
 await act(async () => release(new Response(JSON.stringify({items: [{id: old.id, status: old.status, draftRevision: 1, approvedAt: old.approvedAt}], nextCursor: ''}))));
 expect(screen.getByRole('combobox', {name: 'Saved GPS operations'})).toBeVisible();
 expect(sessionStorage.getItem(key)).toBe(pending);
 expect(screen.getByRole('button', {name: 'Reconcile confirmation'})).toBeVisible();
 expect(screen.queryByRole('button', {name: 'Confirm GPS write'})).not.toBeInTheDocument();
 expect(request.mock.calls.filter(([, init]) => init.method === 'POST')).toHaveLength(1);
});

it('keeps the newly acknowledged operation when an older history detail arrives late', async () => {
 const draft = stagedPreviewDraft(); const preview = savedPreview(); const user = userEvent.setup();
 const current = {id: '99999999-9999-4999-8999-999999999999', plan: preview.plan, digest: preview.digest, status: 'queued', code: '', approvedAt: new Date().toISOString(), attempts: 0, generation: 0, observed: null, verified: false, refreshed: false, noop: false, settled: true, events: []};
 const old = {...current, id: '55555555-5555-4555-8555-555555555555', plan: {...preview.plan, id: '66666666-6666-4666-8666-666666666666'}, status: 'canceled', code: 'DRAFT_CHANGED'};
 let release: (value: Response) => void = () => undefined;
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (path.includes('?draftId=')) {return new Response(JSON.stringify({items: [{id: old.id, status: old.status, draftRevision: 1, approvedAt: old.approvedAt}], nextCursor: ''}));}
  if (init.method === 'POST') {return new Response(JSON.stringify(current));}
  return new Promise<Response>(resolve => {release = resolve;});
 }); vi.stubGlobal('fetch', request);
 render(<WriteOperation owner={'owner'} draft={draft} preview={preview} />);
 await waitFor(() => expect(request.mock.calls.some(([path]) => path.endsWith(`/${old.id}`))).toBe(true));
 await user.click(screen.getByRole('button', {name: 'Confirm GPS write'}));
 expect(await screen.findByText('GPS operation: queued')).toBeVisible();
 await act(async () => release(new Response(JSON.stringify(old))));
 expect(screen.getByText('GPS operation: queued')).toBeVisible();
 expect(screen.queryByText('GPS operation: canceled')).not.toBeInTheDocument();
 expect(screen.queryByRole('button', {name: 'Confirm GPS write'})).not.toBeInTheDocument();
 expect(sessionStorage.length).toBe(0);
});

it('preserves confirmation feedback when the obsolete initial history request fails', async () => {
 const draft = stagedPreviewDraft(); const preview = savedPreview(); const user = userEvent.setup();
 let fail: (error: Error) => void = () => undefined;
 vi.stubGlobal('fetch', vi.fn(async (_path: string, init: RequestInit) => {
  if (init.method === 'POST') {throw new Error('lost acknowledgement');}
  return new Promise<Response>((_resolve, reject) => {fail = reject;});
 }));
 render(<WriteOperation owner={'owner'} draft={draft} preview={preview} />);
 await user.click(screen.getByRole('button', {name: 'Confirm GPS write'}));
 expect(await screen.findByText(/Confirmation acknowledgement unavailable/)).toBeVisible();
 await act(async () => fail(new Error('old history unavailable')));
 expect(screen.getByText(/Confirmation acknowledgement unavailable/)).toBeVisible();
 expect(screen.getByRole('button', {name: 'Reconcile confirmation'})).toBeVisible();
});

it('allows a fresh explicit confirmation after a repeated identity is definitively expired', async () => {
 const draft = stagedPreviewDraft(); const preview = savedPreview(); const user = userEvent.setup();
 const key = `ai-write-confirmation:owner:${draft.id}`;
 const pending = {previewId: preview.plan.id, digest: preview.digest, idempotencyKey: 'a'.repeat(32)};
 sessionStorage.setItem(key, JSON.stringify(pending));
 const fresh = {...preview, digest: 'c'.repeat(64), plan: {...preview.plan, id: '44444444-4444-4444-8444-444444444444'}};
 const current = {id: '99999999-9999-4999-8999-999999999999', plan: fresh.plan, digest: fresh.digest, status: 'queued', code: '', approvedAt: new Date().toISOString(), attempts: 0, generation: 0, observed: null, verified: false, refreshed: false, noop: false, settled: true, events: []};
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (path.includes('?draftId=')) {return new Response(JSON.stringify({items: [], nextCursor: ''}));}
  if (init.method === 'POST') {
   const input = JSON.parse(String(init.body));
   return input.previewId === pending.previewId ? new Response(JSON.stringify({code: 'PREVIEW_EXPIRED'}), {status: 410}) : new Response(JSON.stringify(current));
  }
  return new Response(JSON.stringify({code: 'WRITE_UNAVAILABLE'}), {status: 404});
 }); vi.stubGlobal('fetch', request);
 const view = render(<WriteOperation owner={'owner'} draft={draft} preview={preview} />);
 await waitFor(() => expect(request.mock.calls.some(([path]) => path.includes('/by-key/'))).toBe(true));
 await user.click(screen.getByRole('button', {name: 'Retry same confirmation'}));
 await waitFor(() => expect(request.mock.calls.some(([, init]) => init.method === 'POST')).toBe(true));
 expect(sessionStorage.getItem(key)).toBeNull();
 expect(screen.queryByRole('button', {name: 'Reconcile confirmation'})).not.toBeInTheDocument();
 expect(request.mock.calls.filter(([, init]) => init.method === 'POST')).toHaveLength(1);
 view.rerender(<WriteOperation owner={'owner'} draft={draft} preview={fresh} />);
 await user.click(screen.getByRole('button', {name: 'Confirm GPS write'}));
 expect(await screen.findByText('GPS operation: queued')).toBeVisible();
 const submissions = request.mock.calls.filter(([, init]) => init.method === 'POST').map(([, init]) => JSON.parse(String(init.body)));
 expect(submissions).toHaveLength(2);
 expect(submissions[0]).toEqual(pending);
 expect(submissions[1]).toEqual({previewId: fresh.plan.id, digest: fresh.digest, idempotencyKey: expect.any(String)});
 expect(submissions[1].idempotencyKey).not.toBe(pending.idempotencyKey);
});

it('retains the exact pending identity after missing lookup and an ambiguous repeat failure', async () => {
 const draft = stagedPreviewDraft(); const preview = savedPreview(); const user = userEvent.setup();
 const key = `ai-write-confirmation:owner:${draft.id}`;
 const pending = {previewId: preview.plan.id, digest: preview.digest, idempotencyKey: 'a'.repeat(32)};
 sessionStorage.setItem(key, JSON.stringify(pending));
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (path.includes('?draftId=')) {return new Response(JSON.stringify({items: [], nextCursor: ''}));}
  return new Response(JSON.stringify({code: 'WRITE_UNAVAILABLE'}), {status: init.method === 'POST' ? 503 : 404});
 }); vi.stubGlobal('fetch', request);
 render(<WriteOperation owner={'owner'} draft={draft} preview={preview} />);
 await waitFor(() => expect(request.mock.calls.some(([path]) => path.includes('/by-key/'))).toBe(true));
 expect(sessionStorage.getItem(key)).toBe(JSON.stringify(pending));
 await user.click(screen.getByRole('button', {name: 'Retry same confirmation'}));
 expect(await screen.findByText(/same identity was used/)).toBeVisible();
 expect(sessionStorage.getItem(key)).toBe(JSON.stringify(pending));
 expect(screen.getByRole('button', {name: 'Reconcile confirmation'})).toBeVisible();
 expect(screen.queryByRole('button', {name: 'Confirm GPS write'})).not.toBeInTheDocument();
 const submissions = request.mock.calls.filter(([, init]) => init.method === 'POST');
 expect(submissions).toHaveLength(1);
 expect(JSON.parse(String(submissions[0][1].body))).toEqual(pending);
});
