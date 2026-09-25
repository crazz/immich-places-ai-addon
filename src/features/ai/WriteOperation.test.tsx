import {act, render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import {DraftReview} from './DraftReview';
import {parseReview} from './reviewParser';
import {resultDetail} from './testing/resultDetail';
import {savedPreview, stagedPreviewDraft} from './testing/writePreview';
import {WriteOperation} from './WriteOperation';

afterEach(() => {vi.unstubAllGlobals(); sessionStorage.clear();});

it('keeps unresolved sender completion visible even after intended GPS is observed', async () => {
 const draft = stagedPreviewDraft(); const preview = savedPreview();
 const op = {id: '99999999-9999-4999-8999-999999999999', plan: preview.plan, digest: preview.digest, status: 'succeeded', code: 'GPS_VERIFIED', approvedAt: new Date().toISOString(), attempts: 1, generation: 1, observed: preview.plan.intended, verified: true, refreshed: true, noop: false, settled: false, events: []};
 vi.stubGlobal('fetch', vi.fn(async (path: string) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [{id: op.id, status: op.status, draftRevision: 1, approvedAt: op.approvedAt}], nextCursor: ''} : op))));
 render(<WriteOperation owner={'owner'} draft={draft} preview={null} />);
 expect(await screen.findByText('Prior request completion is unknown. Another GPS write and revision changes remain blocked.')).toBeVisible();
 expect(screen.queryByRole('button', {name: 'Retry GPS write'})).not.toBeInTheDocument();
});

it('shows a definitive disabled rejection without trapping the owner in unknown confirmation recovery', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const preview = savedPreview();
 const request = vi.fn(async (_path: string, init: RequestInit) => new Response(JSON.stringify(init.method === 'POST' ? {code: 'WRITE_DISABLED'} : {items: [], nextCursor: ''}), {status: init.method === 'POST' ? 503 : 200})); vi.stubGlobal('fetch', request);
 render(<WriteOperation owner={'owner'} draft={draft} preview={preview} />);
 await user.click(screen.getByRole('button', {name: 'Confirm GPS write'}));
 expect(await screen.findByText('GPS writing is disabled. Saved review and status remain available.')).toBeVisible();
 expect(screen.queryByRole('button', {name: 'Reconcile confirmation'})).not.toBeInTheDocument();
 expect(sessionStorage.length).toBe(0);
 expect(request.mock.calls.filter(([, init]) => init.method === 'POST')).toHaveLength(1);
});

it('fences late private confirmation replies across account and result changes and preserves manual blocking', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const preview = savedPreview(); const refresh = vi.fn();
 let reply: (value: Response) => void = () => undefined;
 const request = vi.fn(async (_path: string, init: RequestInit) => init.method === 'POST' ? new Promise<Response>(resolve => {reply = resolve;}) : new Response(JSON.stringify({items: [], nextCursor: ''}))); vi.stubGlobal('fetch', request);
 const view = render(<WriteOperation owner={'owner'} draft={draft} preview={preview} disabled onVerified={refresh} />);
 expect(screen.getByRole('button', {name: 'Confirm GPS write'})).toBeDisabled();
 view.rerender(<WriteOperation owner={'owner'} draft={draft} preview={preview} onVerified={refresh} />);
 await user.click(screen.getByRole('button', {name: 'Confirm GPS write'}));
 view.rerender(<WriteOperation owner={'another-owner'} draft={{...draft, id: '66666666-6666-4666-8666-666666666666'}} preview={null} onVerified={refresh} />);
 await act(async () => reply(new Response(JSON.stringify({id: '99999999-9999-4999-8999-999999999999', plan: preview.plan, digest: preview.digest, status: 'succeeded', code: 'GPS_VERIFIED', approvedAt: new Date().toISOString(), attempts: 1, generation: 1, observed: preview.plan.intended, verified: true, refreshed: true, noop: false, settled: true, events: []}))));
 expect(screen.queryByText('GPS operation: succeeded')).not.toBeInTheDocument();
 expect(screen.queryByText(/Approved intended/)).not.toBeInTheDocument();
 expect(refresh).not.toHaveBeenCalled();
 expect(draft.camera).toEqual({latitude: 0, longitude: 12});
 expect(request.mock.calls.filter(([, init]) => init.method === 'POST')).toHaveLength(1);
});

it('confirms by keyboard on HTTP and reconciles a lost acknowledgement using the same key without another POST', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const preview = savedPreview();
 const detail = resultDetail(); detail.entry.draftId = draft.id;
 const op = {id: '99999999-9999-4999-8999-999999999999', plan: preview.plan, digest: preview.digest, status: 'queued', code: '', approvedAt: new Date().toISOString(), attempts: 0, generation: 0, observed: null, verified: false, refreshed: false, noop: false, settled: true, events: []};
 vi.stubGlobal('crypto', {getRandomValues: crypto.getRandomValues.bind(crypto)});
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (path.endsWith('/ai/write-operations') && init.method === 'POST') {throw new Error('lost acknowledgement');}
  if (path.includes('/by-key/')) {return new Response(JSON.stringify(op));}
  if (path.includes('/ai/write-operations?')) {return new Response(JSON.stringify({items: [], nextCursor: ''}));}
  return new Response(JSON.stringify(path.includes('/write-previews') ? preview : draft));
 }); vi.stubGlobal('fetch', request);
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} />);
 await user.click(await screen.findByRole('button', {name: 'Preview exact GPS'}));
 const confirm = await screen.findByRole('button', {name: 'Confirm GPS write'}); confirm.focus(); await user.keyboard('{Enter}');
 expect(await screen.findByText(/Confirmation acknowledgement unavailable/)).toBeVisible();
 const post = request.mock.calls.find(([path, init]) => path.endsWith('/ai/write-operations') && init.method === 'POST')!;
 const input = JSON.parse(String(post[1].body));
 expect(input).toEqual({previewId: preview.plan.id, digest: preview.digest, idempotencyKey: expect.stringMatching(/^[a-f0-9]{32}$/)});
 expect(post[1].redirect).toBe('error');
 await user.click(screen.getByRole('button', {name: 'Reconcile confirmation'}));
 expect(await screen.findByText('GPS operation: queued')).toBeVisible();
 await waitFor(() => expect(request.mock.calls.some(([path]) => path.endsWith(`/by-key/${input.idempotencyKey}`))).toBe(true));
 expect(request.mock.calls.filter(([, init]) => init.method === 'POST')).toHaveLength(2);
 expect(request.mock.calls.filter(([path, init]) => path.endsWith('/ai/write-operations') && init.method === 'POST')).toHaveLength(1);
});

it('polls only private reads until durable verification and stops after the bounded read window', async () => {
 vi.useFakeTimers();
 try {
  const draft = stagedPreviewDraft(); const preview = savedPreview(); const refresh = vi.fn();
  const op = {id: '99999999-9999-4999-8999-999999999999', plan: preview.plan, digest: preview.digest, status: 'queued', code: '', approvedAt: new Date().toISOString(), attempts: 0, generation: 0, observed: null, verified: false, refreshed: false, noop: false, settled: true, events: []};
  let reads = 0;
  const request = vi.fn<(path: string, init: RequestInit) => Promise<Response>>(async (path: string) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [{id: op.id, status: op.status, draftRevision: 1, approvedAt: op.approvedAt}], nextCursor: ''} : ++reads === 1 ? op : {...op, status: 'succeeded', verified: true, refreshed: true, observed: preview.plan.intended}))); vi.stubGlobal('fetch', request);
  render(<WriteOperation owner={'owner'} draft={draft} preview={null} onVerified={refresh} />);
  await act(async () => {await vi.advanceTimersByTimeAsync(0);});
  expect(screen.getByText('GPS operation: queued')).toBeVisible();
  await act(async () => {await vi.advanceTimersByTimeAsync(2_000);});
  expect(screen.getByText('GPS operation: succeeded')).toBeVisible();
  expect(refresh).toHaveBeenCalledTimes(1);
  const calls = request.mock.calls.length;
  await act(async () => {await vi.advanceTimersByTimeAsync(60_000);});
  expect(request).toHaveBeenCalledTimes(calls);
 } finally {vi.useRealTimers();}
});

it('checks an unresolved operation without a new mutation and reports upstream verification with pending local refresh distinctly', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const preview = savedPreview();
 const op = {id: '99999999-9999-4999-8999-999999999999', plan: preview.plan, digest: preview.digest, status: 'verifying', code: 'RECONCILIATION_REQUIRED', approvedAt: new Date().toISOString(), attempts: 1, generation: 1, observed: null, verified: false, refreshed: false, noop: false, settled: true, events: []};
 const request = vi.fn(async (path: string, init: RequestInit) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [{id: op.id, status: op.status, draftRevision: 1, approvedAt: op.approvedAt}], nextCursor: ''} : init.method === 'POST' ? {...op, verified: true, code: 'LOCAL_REFRESH_PENDING', observed: preview.plan.intended} : op))); vi.stubGlobal('fetch', request);
 render(<WriteOperation owner={'owner'} draft={draft} preview={null} disabled />);
 expect(await screen.findByText('GPS operation: verifying')).toBeVisible();
 await user.click(screen.getByRole('button', {name: 'Check status'}));
 expect(await screen.findByText('GPS verified upstream; local refresh pending.')).toBeVisible();
 expect(request.mock.calls.filter(([, init]) => init.method === 'POST')).toHaveLength(1);
 expect(request.mock.calls.find(([, init]) => init.method === 'POST')?.[0]).toContain(`/${op.id}/reconcile`);
 expect(screen.queryByRole('button', {name: 'Retry GPS write'})).not.toBeInTheDocument();
});

it('requires an explicit generation-bound retry and refreshes the catalog only after verified publication', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const preview = savedPreview();
 const op = {id: '99999999-9999-4999-8999-999999999999', plan: preview.plan, digest: preview.digest, status: 'retryable', code: 'RETRY_AVAILABLE', approvedAt: new Date().toISOString(), attempts: 1, generation: 4, observed: preview.plan.before, verified: false, refreshed: false, noop: false, settled: true, events: []};
 const refresh = vi.fn();
 const request = vi.fn<(path: string, init: RequestInit) => Promise<Response>>(async (path: string) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [{id: op.id, status: op.status, draftRevision: 1, approvedAt: op.approvedAt}], nextCursor: ''} : path.endsWith('/retry') ? {...op, status: 'succeeded', verified: true, refreshed: true, observed: preview.plan.intended, attempts: 2, generation: 5} : op))); vi.stubGlobal('fetch', request);
 const view = render(<WriteOperation owner={'owner'} draft={draft} preview={null} disabled onVerified={refresh} />);
 expect(await screen.findByText('GPS operation: retryable')).toBeVisible();
 expect(screen.getByRole('button', {name: 'Retry GPS write'})).toBeDisabled();
 expect(refresh).not.toHaveBeenCalled();
 view.rerender(<WriteOperation owner={'owner'} draft={draft} preview={null} onVerified={refresh} />);
 await user.click(screen.getByRole('button', {name: 'Retry GPS write'}));
 expect(await screen.findByText('GPS operation: succeeded')).toBeVisible();
 expect(JSON.parse(String(request.mock.calls.find(([path]) => path.endsWith('/retry'))?.[1]?.body))).toEqual({generation: 4});
 await waitFor(() => expect(refresh).toHaveBeenCalledTimes(1));
 expect(screen.queryByRole('button', {name: 'Retry GPS write'})).not.toBeInTheDocument();
});

it('repeats the retained confirmation identity explicitly after a lost response', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const preview = savedPreview(); let sends = 0;
 const op = {id: '99999999-9999-4999-8999-999999999999', plan: preview.plan, digest: preview.digest, status: 'queued', code: '', approvedAt: new Date().toISOString(), attempts: 0, generation: 0, observed: null, verified: false, refreshed: false, noop: false, settled: true, events: []};
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (init.method === 'POST' && ++sends === 1) {throw new Error('lost');}
  return new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : op));
 }); vi.stubGlobal('fetch', request);
 render(<WriteOperation owner={'owner'} draft={draft} preview={preview} />);
 await user.click(screen.getByRole('button', {name: 'Confirm GPS write'}));
 await screen.findByText(/Confirmation acknowledgement unavailable/);
 await user.click(screen.getByRole('button', {name: 'Retry same confirmation'}));
 expect(await screen.findByText('GPS operation: queued')).toBeVisible();
 const posts = request.mock.calls.filter(([, init]) => init.method === 'POST');
 expect(posts).toHaveLength(2); expect(posts[1][1].body).toBe(posts[0][1].body);
});

it('pages through retained GPS operations without dispatching or losing the inspected operation', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const preview = savedPreview();
 const op = {id: '99999999-9999-4999-8999-999999999999', plan: preview.plan, digest: preview.digest, status: 'succeeded', code: 'GPS_VERIFIED', approvedAt: new Date().toISOString(), attempts: 1, generation: 1, observed: preview.plan.intended, verified: true, refreshed: true, noop: false, settled: true, events: []};
 const older = {...op, id: '88888888-8888-4888-8888-888888888888', status: 'failed', verified: false, refreshed: false};
 const summary = (value: typeof op): {id: string; status: string; draftRevision: number; approvedAt: string} => ({id: value.id, status: value.status, draftRevision: 1, approvedAt: value.approvedAt});
 const request = vi.fn<(path: string, init: RequestInit) => Promise<Response>>(async (path: string) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [summary(path.includes('&before=') ? older : op)], nextCursor: path.includes('&before=') ? '' : op.id} : path.endsWith(older.id) ? older : op))); vi.stubGlobal('fetch', request);
 render(<WriteOperation owner={'owner'} draft={draft} preview={null} />);
 expect(await screen.findByText('GPS operation: succeeded')).toBeVisible();
 await user.click(screen.getByRole('button', {name: 'More saved GPS operations'}));
 expect(screen.getByText('GPS operation: succeeded')).toBeVisible();
 await user.selectOptions(screen.getByLabelText('Saved GPS operations'), older.id);
 expect(await screen.findByText('GPS operation: failed')).toBeVisible();
 expect(screen.queryByRole('button', {name: 'More saved GPS operations'})).not.toBeInTheDocument();
 expect(request.mock.calls.every(([, init]) => init.method === 'GET')).toBe(true);
 expect(request.mock.calls.some(([path]) => path.includes(`&before=${op.id}`))).toBe(true);
});
