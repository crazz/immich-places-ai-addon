import {render, screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import {DraftReview} from './DraftReview';
import {parseReview} from './reviewParser';
import {mirrorOperation, mirroredDraft} from './testing/mirror';
import {resultDetail} from './testing/resultDetail';

afterEach(() => {vi.unstubAllGlobals(); sessionStorage.clear();});

it('M20 retains standard success and independent mirror recovery alongside local direction and languages', async () => {
 const user = userEvent.setup(); const draft = mirroredDraft(); const op = mirrorOperation(); const detail = resultDetail(); detail.entry.draftId = draft.id; const refresh = vi.fn();
 const request = vi.fn<(path: string, init: RequestInit) => Promise<Response>>(async path => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [{id: op.id, status: op.status, draftRevision: 1, approvedAt: op.approvedAt}], nextCursor: ''} : path.includes('/write-operations/') ? op : draft))); vi.stubGlobal('fetch', request);
 const view = render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} onVerified={refresh} />);
 expect(await screen.findByText('Metadata step: retryable')).toBeVisible();
 const standard = within(screen.getByRole('region', {name: `Outcome for ${draft.assetId}`})); expect(standard.getByText('Target status: succeeded')).toBeVisible();
 expect(screen.getByRole('button', {name: 'Retry metadata only'})).toBeEnabled(); expect(screen.queryByRole('button', {name: `Retry GPS for ${draft.assetId}`})).not.toBeInTheDocument();
 expect(screen.getByLabelText('Description uk')).toHaveValue(draft.descriptions[0].text); expect(screen.getByLabelText('Heading degrees')).toHaveValue(null); await waitFor(() => expect(refresh).toHaveBeenCalledTimes(1));
 view.unmount(); op.mirror = {...op.mirror, status: 'verifying', code: 'METADATA_UNAVAILABLE', verified: true, settled: false, observed: {present: true, value: op.plan.mirror.value}};
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} onVerified={refresh} />);
 expect(await screen.findByText('Metadata step: verifying')).toBeVisible(); expect(screen.getByText('Metadata sender completion is unresolved; another metadata write is not permitted.')).toBeVisible();
 expect(screen.queryByRole('button', {name: 'Retry metadata only'})).not.toBeInTheDocument(); expect(screen.getByRole('button', {name: 'Check metadata status'})).toBeEnabled();
 expect(screen.getByLabelText('Description uk')).toHaveValue(draft.descriptions[0].text); expect(refresh).toHaveBeenCalledTimes(1);
 await user.click(screen.getByText('Exact metadata before and after')); expect(screen.getByLabelText('Proposed metadata').textContent).toContain('Місто');
 expect(request.mock.calls.every(([, init]) => init.method === 'GET')).toBe(true);
});

it('replays only the retained metadata target and generation after a lost acknowledgement and reload', async () => {
 const {WriteOperation} = await import('./WriteOperation'); const user = userEvent.setup(); const draft = mirroredDraft(); const op = mirrorOperation(); let retries = 0;
 const done = {...op, status: 'succeeded', verified: true, mirror: {...op.mirror, status: 'succeeded', attempts: 2, generation: 4, verified: true, observed: {present: true, value: op.plan.mirror.value}}};
 const request = vi.fn<(path: string, init: RequestInit) => Promise<Response>>(async path => {
  if (path.endsWith('/metadata/retry') && ++retries === 1) {throw new Error('lost acknowledgement');}
  return new Response(JSON.stringify(path.includes('?draftId=') ? {items: [{id: op.id, status: op.status, draftRevision: 1, approvedAt: op.approvedAt}], nextCursor: ''} : path.endsWith('/metadata/retry') ? done : op));
 }); vi.stubGlobal('fetch', request);
 const view = render(<WriteOperation owner={'owner'} draft={draft} preview={null} />);
 await user.click(await screen.findByRole('button', {name: 'Retry metadata only'}));
 expect(await screen.findByText('Metadata retry acknowledgement unavailable. Check status or repeat the same metadata generation.')).toBeVisible(); view.unmount();
 render(<WriteOperation owner={'owner'} draft={{...draft, revision: 2}} preview={null} disabled manualPendingIDs={[draft.assetId]} />);
 await user.click(await screen.findByRole('button', {name: 'Repeat metadata retry generation 3'}));
 expect(await screen.findByText('Metadata step: succeeded')).toBeVisible();
 const posts = request.mock.calls.filter(([, init]) => init.method === 'POST'); expect(posts).toHaveLength(2);
 expect(posts.every(([path]) => path.endsWith(`/${op.id}/targets/${draft.assetId}/metadata/retry`))).toBe(true); expect(posts.map(([, init]) => JSON.parse(String(init.body)))).toEqual([{generation: 3}, {generation: 3}]);
 expect(screen.getByText('Target status: succeeded')).toBeVisible(); expect(screen.queryByRole('button', {name: `Retry GPS for ${draft.assetId}`})).not.toBeInTheDocument();
});

it('M21 fences late mirror status after newer history and late confirmation after an account change', async () => {
 const {act, fireEvent} = await import('@testing-library/react'); const {WriteOperation} = await import('./WriteOperation'); vi.useFakeTimers();
 try {
  const draft = mirroredDraft(); const op = mirrorOperation(); op.mirror.status = 'verifying'; op.mirror.settled = false;
  const newer = {...op, id: '55555555-5555-4555-8555-555555555555', mirror: {...op.mirror, status: 'failed', code: 'NEWER_PRIVATE_OUTCOME', settled: true}};
  let reads = 0; let releaseStatus: (response: Response) => void = () => undefined; let releaseConfirmation: (response: Response) => void = () => undefined;
  const request = vi.fn(async (path: string, init: RequestInit) => {
   if (init.method === 'POST') {return new Promise<Response>(resolve => {releaseConfirmation = resolve;});}
   if (path.includes('?draftId=')) {return new Response(JSON.stringify({items: [op, newer].map(item => ({id: item.id, status: item.status, draftRevision: 1, approvedAt: item.approvedAt})), nextCursor: ''}));}
   if (path.endsWith(newer.id)) {return new Response(JSON.stringify(newer));}
   if (++reads > 1) {return new Promise<Response>(resolve => {releaseStatus = resolve;});}
   return new Response(JSON.stringify(op));
  }); vi.stubGlobal('fetch', request);
  const preview = {plan: {...op.plan, id: '66666666-6666-4666-8666-666666666666'}, digest: op.digest, status: 'usable' as const, diff: 'changed' as const};
  const view = render(<WriteOperation owner={'owner'} draft={draft} preview={preview} />);
  await act(async () => {await vi.advanceTimersByTimeAsync(0);});
  await act(async () => {await vi.advanceTimersByTimeAsync(2000);});
  expect(reads).toBe(2);
  fireEvent.change(screen.getByLabelText('Saved GPS operations'), {target: {value: newer.id}}); await act(async () => {await vi.advanceTimersByTimeAsync(0);});
  expect(screen.getByText('Metadata outcome: NEWER_PRIVATE_OUTCOME')).toBeVisible();
  await act(async () => releaseStatus(new Response(JSON.stringify(op))));
  expect(screen.getByText('Metadata outcome: NEWER_PRIVATE_OUTCOME')).toBeVisible();
  fireEvent.click(screen.getByRole('button', {name: 'Confirm GPS write'})); await act(async () => {await vi.advanceTimersByTimeAsync(0);});
  const retained = sessionStorage.getItem(`ai-write-confirmation:owner:${draft.id}`); expect(retained).not.toBeNull();
  view.rerender(<WriteOperation owner={'other'} draft={{...draft, id: '77777777-7777-4777-8777-777777777777'}} preview={null} />);
  await act(async () => releaseConfirmation(new Response(JSON.stringify({...op, plan: preview.plan}))));
  expect(screen.queryByText('Metadata outcome: NEWER_PRIVATE_OUTCOME')).not.toBeInTheDocument(); expect(screen.queryByLabelText('Proposed metadata')).not.toBeInTheDocument();
  expect(sessionStorage.getItem(`ai-write-confirmation:owner:${draft.id}`)).toBe(retained);
 } finally {vi.useRealTimers();}
});

it('reconciles only the unresolved metadata generation and preserves successful standard fields', async () => {
 const {WriteOperation} = await import('./WriteOperation'); const user = userEvent.setup(); const draft = mirroredDraft(); const op = mirrorOperation(); op.mirror.status = 'verifying'; op.mirror.settled = false;
 const reconciled = {...op, mirror: {...op.mirror, generation: 4, status: 'failed', code: 'METADATA_UNAVAILABLE', settled: true}};
 const request = vi.fn(async (path: string, init: RequestInit) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [{id: op.id, status: op.status, draftRevision: 1, approvedAt: op.approvedAt}], nextCursor: ''} : init.method === 'POST' ? reconciled : op))); vi.stubGlobal('fetch', request);
 render(<WriteOperation owner={'owner'} draft={{...draft, revision: 2}} preview={null} disabled manualPendingIDs={[draft.assetId]} />);
 await user.click(await screen.findByRole('button', {name: 'Check metadata status'}));
 expect(await screen.findByText('Metadata step: failed')).toBeVisible(); expect(screen.getByText('Target status: succeeded')).toBeVisible(); expect(screen.getByText(/current draft revision 2 is not marked saved/)).toBeVisible();
 const posts = request.mock.calls.filter(([, init]) => init.method === 'POST'); expect(posts).toHaveLength(1);
 expect(posts[0][0]).toContain(`/${op.id}/targets/${draft.assetId}/metadata/reconcile`); expect(JSON.parse(String(posts[0][1].body))).toEqual({generation: 3});
 expect(screen.queryByRole('button', {name: 'Retry metadata only'})).not.toBeInTheDocument();
});

it('shows prior verification separately from a current metadata conflict', async () => {
 const {MirrorOutcome} = await import('./MirrorOutcome'); const op = mirrorOperation(); op.mirror = {...op.mirror, status: 'conflict', code: 'METADATA_CONFLICT', verified: true, observed: {present: false}};
 render(<MirrorOutcome operation={op} />);
 expect(screen.getByText('Metadata step: conflict')).toBeVisible();
 expect(screen.getByText('Approved metadata was previously verified; the latest observation conflicts with it.')).toBeVisible();
 expect(screen.queryByText('Approved metadata observed upstream; this does not prove native display or file updates.')).not.toBeInTheDocument();
 expect(screen.queryByRole('button', {name: 'Retry metadata only'})).not.toBeInTheDocument();
});

it('permits metadata-only recovery after standard success without resolving newly pending manual GPS', async () => {
 const user = userEvent.setup(); const draft = mirroredDraft(); const op = mirrorOperation(); const detail = resultDetail(); detail.entry.draftId = draft.id; const pending = [draft.assetId];
 const request = vi.fn<(path: string, init: RequestInit) => Promise<Response>>(async path => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [{id: op.id, status: op.status, draftRevision: 1, approvedAt: op.approvedAt}], nextCursor: ''} : path.includes('/write-operations/') ? op : draft))); vi.stubGlobal('fetch', request);
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} hasManualConflict manualPendingIDs={pending} />);
 const retry = await screen.findByRole('button', {name: 'Retry metadata only'}); expect(retry).toBeEnabled(); await user.click(retry);
 await waitFor(() => expect(request.mock.calls.filter(([, init]) => init.method === 'POST')).toHaveLength(1));
 const post = request.mock.calls.find(([, init]) => init.method === 'POST')!; expect(post[0]).toContain(`/${op.id}/targets/${draft.assetId}/metadata/retry`); expect(JSON.parse(String(post[1].body))).toEqual({generation: 3});
 expect(pending).toEqual([draft.assetId]); expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeDisabled();
});


it('reports the completed combined operation without a pending parent-code fallback', async () => {
 const {WriteOperation} = await import('./WriteOperation'); const draft = mirroredDraft(); const op = mirrorOperation();
 op.status = 'succeeded'; op.code = ''; op.verified = true;
 op.mirror = {...op.mirror, status: 'succeeded', code: 'METADATA_VERIFIED', verified: true, observed: {present: true, value: op.plan.mirror.value}};
 vi.stubGlobal('fetch', vi.fn(async (path: string) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [{id: op.id, status: op.status, draftRevision: 1, approvedAt: op.approvedAt}], nextCursor: ''} : op))));
 render(<WriteOperation owner={'owner'} draft={draft} preview={null} />);
 expect(await screen.findByText('Metadata step: succeeded')).toBeVisible();
 expect(screen.queryByText('Outcome: awaiting execution')).not.toBeInTheDocument();
 expect(screen.getByText('Combined operation: succeeded')).toBeVisible();
 expect(screen.getByText('Outcome: succeeded')).toBeVisible();
});
