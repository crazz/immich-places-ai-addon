import {act, render, screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import {stackOperation,stagedPreviewDraft} from './testing/writePreview';
import {WriteOperation} from './WriteOperation';

afterEach(() => {vi.unstubAllGlobals(); sessionStorage.clear();});

it('S19 P16 retains distinct target outcomes after reload and fences delayed private history', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const op = stackOperation();
 const preview = {plan: op.plan, digest: op.digest, status: 'usable' as const, diff: 'changed' as const};
 const old = {...op, id: '55555555-5555-4555-8555-555555555555', status: 'failed'};
 let release: (value: Response) => void = () => undefined; let shouldDelay = true;
 const page = {items: [{id: op.id, status: op.status, draftRevision: draft.revision, approvedAt: op.approvedAt}], nextCursor: ''};
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (path.includes('?draftId=') && shouldDelay) {return new Promise<Response>(resolve => {release = resolve;});}
  return new Response(JSON.stringify(path.includes('?draftId=') ? page : init.method === 'POST' ? op : op));
 }); vi.stubGlobal('fetch', request);
 const view = render(<WriteOperation owner={'owner'} draft={draft} preview={preview} />);
 await user.click(screen.getByRole('button', {name: 'Confirm GPS write'}));
 expect(await screen.findByText('Stack operation: partial')).toBeVisible();
 for (const target of op.targets) {expect(within(screen.getByRole('region', {name: `Outcome for ${target.assetId}`})).getByText(`Target status: ${target.status}`)).toBeVisible();}
 expect(screen.getByText('GPS verified unchanged; no write sent for this target.')).toBeVisible();
 expect(screen.getByText('Sender completion unresolved; target exclusion is retained.')).toBeVisible();
 await act(async () => release(new Response(JSON.stringify({items: [{id: old.id, status: old.status, draftRevision: 1, approvedAt: old.approvedAt}], nextCursor: ''}))));
 expect(screen.getByText('Stack operation: partial')).toBeVisible();
 shouldDelay = false; view.unmount();
 const restored = render(<WriteOperation owner={'owner'} draft={draft} preview={null} />);
 expect(await screen.findByText('Stack operation: partial')).toBeVisible();
 expect(screen.getByRole('region', {name: `Outcome for ${op.targets[0].assetId}`})).toBeVisible();
 restored.unmount(); shouldDelay = true;
 const switching = render(<WriteOperation owner={'owner'} draft={draft} preview={null} />);
 await waitFor(() => expect(request.mock.calls.filter(([path]) => path.includes('?draftId=')).length).toBe(3));
 const releasePrevious = release;
 switching.rerender(<WriteOperation owner={'other'} draft={{...draft, id: '66666666-6666-4666-8666-666666666666'}} preview={null} />);
 await act(async () => releasePrevious(new Response(JSON.stringify(page))));
 expect(screen.queryByText('Stack operation: partial')).not.toBeInTheDocument();
 expect(screen.queryByRole('region', {name: `Outcome for ${op.targets[0].assetId}`})).not.toBeInTheDocument();
 expect(request.mock.calls.filter(([, init]) => init.method === 'POST')).toHaveLength(1);
});

it('replays only the exact target retry generation after a lost acknowledgement while preserving successful siblings', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const op = stackOperation();
 op.targets[1] = {...op.targets[1], status: 'retryable', code: 'RETRY_AVAILABLE', generation: 4, settled: true};
 const id = op.targets[1].assetId; let retries = 0;
 const completed = {...op, targets: op.targets.map(target => target.assetId === id ? {...target, status: 'succeeded', attempts: 2, generation: 5, verified: true, refreshed: true, fields: [{field: 'gps', status: 'verified', wasVerified: true, gps: op.plan.intended}]} : target)};
 const request = vi.fn<(path: string, init: RequestInit) => Promise<Response>>(async (path: string) => {
  if (path.endsWith('/retry') && ++retries === 1) {throw new Error('lost acknowledgement');}
  return new Response(JSON.stringify(path.includes('?draftId=') ? {items: [{id: op.id, status: op.status, draftRevision: 1, approvedAt: op.approvedAt}], nextCursor: ''} : path.endsWith('/retry') ? completed : op));
 }); vi.stubGlobal('fetch', request);
 const view = render(<WriteOperation owner={'owner'} draft={draft} preview={null} />);
 const retry = await screen.findByRole('button', {name: `Retry GPS for ${id}`}); await user.click(retry);
 expect(await screen.findByText(`Retry acknowledgement unavailable for ${id}. Reconcile status or repeat the same target generation.`)).toBeVisible();
 expect(screen.queryByRole('button', {name: `Retry GPS for ${op.targets[0].assetId}`})).not.toBeInTheDocument();
 view.rerender(<WriteOperation owner={'owner'} draft={{...draft, revision: 2}} preview={null} disabled />);
 await user.click(await screen.findByRole('button', {name: `Repeat retry generation 4 for ${id}`}));
 expect(await within(screen.getByRole('region', {name: `Outcome for ${id}`})).findByText('Target status: succeeded')).toBeVisible();
 const posts = request.mock.calls.filter(([path]) => path.endsWith('/retry'));
 expect(posts).toHaveLength(2);
 expect(posts.every(([path]) => path.endsWith(`/${op.id}/targets/${id}/retry`))).toBe(true);
 expect(posts.map(call => JSON.parse(String(call[1].body)))).toEqual([{generation: 4}, {generation: 4}]);
 expect(screen.getByText('GPS verified unchanged; no write sent for this target.')).toBeVisible();
});

it('reconciles an exact unresolved target and polls active targets despite a partial aggregate', async () => {
 vi.useFakeTimers();
 try {
  const {fireEvent} = await import('@testing-library/react'); const draft = stagedPreviewDraft(); const op = stackOperation(); const refresh = vi.fn();
  const unresolved = op.targets[4]; let reads = 0;
  const updated = {...op, settled: true, targets: op.targets.map(target => target === unresolved ? {...target, status: 'succeeded', generation: 2, verified: true, refreshed: true, settled: true, fields: [{field: 'gps', status: 'verified', wasVerified: true, gps: op.plan.intended}]} : target)};
  const request = vi.fn(async (path: string, init: RequestInit) => {
   if (path.includes('?draftId=')) {return new Response(JSON.stringify({items: [{id: op.id, status: op.status, draftRevision: 1, approvedAt: op.approvedAt}], nextCursor: ''}));}
   if (init.method === 'POST') {return new Response(JSON.stringify({...op, targets: op.targets.map(target => target === unresolved ? {...target, generation: 2} : target)}));}
   return new Response(JSON.stringify(++reads === 1 ? op : updated));
  }); vi.stubGlobal('fetch', request);
  render(<WriteOperation owner={'owner'} draft={draft} preview={null} disabled onVerified={refresh} />);
  await act(async () => {await vi.advanceTimersByTimeAsync(0);});
  expect(refresh).toHaveBeenCalledTimes(1);
  fireEvent.click(screen.getByRole('button', {name: `Check status for ${unresolved.assetId}`}));
  await act(async () => {await vi.advanceTimersByTimeAsync(0);});
  const post = request.mock.calls.find(([, init]) => init.method === 'POST')!;
  expect(post[0]).toContain(`/${op.id}/targets/${unresolved.assetId}/reconcile`); expect(JSON.parse(String(post[1].body))).toEqual({generation: 1});
  await act(async () => {await vi.advanceTimersByTimeAsync(2_000);});
  expect(within(screen.getByRole('region', {name: `Outcome for ${unresolved.assetId}`})).getByText('Target status: succeeded')).toBeVisible();
  expect(refresh).toHaveBeenCalledTimes(2);
  await act(async () => {await vi.advanceTimersByTimeAsync(30_000);});
  expect(refresh).toHaveBeenCalledTimes(2); expect(request.mock.calls.filter(([, init]) => init.method === 'POST')).toHaveLength(1);
 } finally {vi.useRealTimers();}
});

it('fences an obsolete revision history reply after a newer target reconciliation', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const op = stackOperation(); const target = op.targets[4];
 const updated = {...op, targets: op.targets.map(item => item === target ? {...item, generation: 5, status: 'failed', settled: true} : item)};
 const page = {items: [{id: op.id, status: op.status, draftRevision: 1, approvedAt: op.approvedAt}], nextCursor: ''};
 let historyReads = 0; let release: (value: Response) => void = () => undefined;
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (path.includes('?draftId=') && ++historyReads > 1) {return new Promise<Response>(resolve => {release = resolve;});}
  return new Response(JSON.stringify(path.includes('?draftId=') ? page : init.method === 'POST' ? updated : op));
 }); vi.stubGlobal('fetch', request);
 const view = render(<WriteOperation owner={'owner'} draft={draft} preview={null} />);
 await screen.findByText('Stack operation: partial');
 view.rerender(<WriteOperation owner={'owner'} draft={{...draft, revision: 2}} preview={null} />);
 await user.click(screen.getByRole('button', {name: `Check status for ${target.assetId}`}));
 const row = within(screen.getByRole('region', {name: `Outcome for ${target.assetId}`}));
 expect(await row.findByText('Attempts 1 of 2 · Generation 5')).toBeVisible();
 await act(async () => release(new Response(JSON.stringify(page))));
 expect(row.getByText('Attempts 1 of 2 · Generation 5')).toBeVisible();
 expect(row.getByText('Target status: failed')).toBeVisible();
});

it('retains a target retry identity and verified refresh evidence through reload', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const op = stackOperation(); const refresh = vi.fn();
 op.targets[1] = {...op.targets[1], status: 'retryable', generation: 7, settled: true}; const target = op.targets[1];
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (init.method === 'POST') {throw new Error('lost acknowledgement');}
  return new Response(JSON.stringify(path.includes('?draftId=') ? {items: [{id: op.id, status: op.status, draftRevision: 1, approvedAt: op.approvedAt}], nextCursor: ''} : op));
 }); vi.stubGlobal('fetch', request);
 const view = render(<WriteOperation owner={'owner'} draft={draft} preview={null} onVerified={refresh} />);
 await user.click(await screen.findByRole('button', {name: `Retry GPS for ${target.assetId}`}));
 await screen.findByText(/Retry acknowledgement unavailable/); expect(refresh).toHaveBeenCalledTimes(1);
 view.unmount();
 render(<WriteOperation owner={'owner'} draft={draft} preview={null} disabled onVerified={refresh} />);
 await user.click(await screen.findByRole('button', {name: `Repeat retry generation 7 for ${target.assetId}`}));
 await screen.findByText(/Retry acknowledgement unavailable/); expect(refresh).toHaveBeenCalledTimes(1);
 const sends = request.mock.calls.filter(([, init]) => init.method === 'POST');
 expect(sends).toHaveLength(2); expect(sends[0][0]).toBe(sends[1][0]); expect(sends[0][1].body).toBe(sends[1][1].body);
});

it('blocks an uncertain target retry replay when manual GPS becomes pending while preserving read-only reconciliation', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const op = stackOperation(); const target = op.targets[1]; target.status = 'retryable';
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (init.method === 'POST') {throw new Error('lost acknowledgement');}
  return new Response(JSON.stringify(path.includes('?draftId=') ? {items: [{id: op.id, status: op.status, draftRevision: 1, approvedAt: op.approvedAt}], nextCursor: ''} : op));
 }); vi.stubGlobal('fetch', request);
 const view = render(<WriteOperation owner={'owner'} draft={draft} preview={null} />);
 await user.click(await screen.findByRole('button', {name: `Retry GPS for ${target.assetId}`}));
 await screen.findByText(/Retry acknowledgement unavailable/);
 view.rerender(<WriteOperation owner={'owner'} draft={draft} preview={null} manualPendingIDs={[target.assetId]} />);
 expect(screen.getByRole('button', {name: `Repeat retry generation 1 for ${target.assetId}`})).toBeDisabled();
 expect(screen.getByRole('button', {name: `Check status for ${target.assetId}`})).toBeEnabled();
 expect(request.mock.calls.filter(([, init]) => init.method === 'POST')).toHaveLength(1);
});

it('identifies the full target matrix before confirmation and submits only immutable preview identity', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const op = stackOperation();
 const preview = {plan: op.plan, digest: op.digest, status: 'usable' as const, diff: 'changed' as const};
 const request = vi.fn<(path: string, init: RequestInit) => Promise<Response>>(async (path: string) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : op))); vi.stubGlobal('fetch', request);
 render(<WriteOperation owner={'owner'} draft={draft} preview={preview} />);
 expect(screen.getByText('Confirming authorizes only the exact target and field matrix shown above. Manual pending choices are preserved.')).toBeVisible();
 await user.click(screen.getByRole('button', {name: 'Confirm GPS write'}));
 expect(await screen.findByText('Stack operation: partial')).toBeVisible();
 const post = request.mock.calls.find(([path]) => path.endsWith('/write-operations'))!;
 expect(JSON.parse(String(post[1].body))).toEqual({previewId: preview.plan.id, digest: preview.digest, idempotencyKey: expect.stringMatching(/^[a-f0-9]{32}$/)});
});
