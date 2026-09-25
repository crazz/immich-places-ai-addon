import {act, fireEvent, render, screen} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {DraftReview} from './DraftReview';
import {parseReview} from './reviewParser';
import {resultDetail} from './testing/resultDetail';
import {savedPreview, stagedPreviewDraft} from './testing/writePreview';
import {WritePreview} from './WritePreview';

afterEach(() => {vi.unstubAllGlobals(); sessionStorage.clear();});

it('retires the stored reference before a replacement comparison fails', async () => {
 const draft = stagedPreviewDraft(); const preview = savedPreview();
 const key = `ai-write-preview:owner:${draft.id}:${draft.revision}`;
 const request = vi.fn().mockResolvedValueOnce(new Response(JSON.stringify(preview))).mockResolvedValueOnce(new Response(JSON.stringify({code: 'IMMICH_CONFLICT'}), {status: 409}));
 vi.stubGlobal('fetch', request);
 render(<WritePreview owner={'owner'} draft={draft} />);
 fireEvent.click(screen.getByRole('button', {name: 'Preview exact GPS'}));
 expect(await screen.findByText('Preview status: usable')).toBeVisible();
 expect(sessionStorage.getItem(key)).toBe(preview.plan.id);
 fireEvent.click(screen.getByRole('button', {name: 'Preview exact GPS'}));
 expect(await screen.findByText(/GPS changed/)).toBeVisible();
 expect(sessionStorage.getItem(key)).toBeNull();
 expect(screen.queryByText('Preview status: usable')).not.toBeInTheDocument();
});

it('shows an exact numeric GPS comparison for a coarse staged decision without a mutation control', async () => {
 const draft = stagedPreviewDraft();
 const detail = resultDetail(); detail.entry.draftId = draft.id;
 const request = vi.fn().mockImplementation(async (path: string) => new Response(JSON.stringify(path.includes('/write-previews') ? savedPreview() : draft), {status: 200}));
 vi.stubGlobal('fetch', request);
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} />);
 fireEvent.click(await screen.findByRole('button', {name: 'Preview exact GPS'}));
 expect(await screen.findByLabelText('Before latitude')).toHaveValue('0');
 expect(screen.getByLabelText('Before longitude')).toHaveValue('absent');
 expect(screen.getByLabelText('Proposed latitude')).toHaveValue('0');
 expect(screen.getByLabelText('Proposed longitude')).toHaveValue('12');
 expect(screen.getByText('GPS comparison: changed')).toBeVisible();
 expect(screen.getByText('Preview status: usable')).toBeVisible();
 expect(screen.getByText(/Single photo · GPS only/)).toBeVisible();
 expect(screen.queryByRole('button', {name: /Confirm|Write to Immich/})).not.toBeInTheDocument();
 const calls = request.mock.calls.filter(([path]) => String(path).includes('write-previews'));
 expect(calls).toHaveLength(1);
 expect(JSON.parse(calls[0][1].body)).toEqual({draftId: draft.id, draftRevision: draft.revision});
 expect(request.mock.calls.some(([path]) => String(path).includes('/location'))).toBe(false);
});

it('fences late private replies across result changes, account changes and a new manual overlap', async () => {
 const draft = stagedPreviewDraft();
 let reply: (value: Response) => void = () => {};
 vi.stubGlobal('fetch', vi.fn(async () => new Promise<Response>(resolve => {reply = resolve;})));
 const view = render(<WritePreview owner={'owner'} draft={draft} />);
 fireEvent.click(screen.getByRole('button', {name: 'Preview exact GPS'}));
 view.rerender(<WritePreview owner={'owner'} draft={{...draft, id: '66666666-6666-4666-8666-666666666666'}} />);
 await act(async () => reply(new Response(JSON.stringify(savedPreview()))));
 expect(screen.queryByLabelText('Before latitude')).not.toBeInTheDocument();
 view.rerender(<WritePreview owner={'owner'} draft={draft} />);
 fireEvent.click(screen.getByRole('button', {name: 'Preview exact GPS'}));
 view.rerender(<WritePreview owner={'another-owner'} draft={draft} />);
 await act(async () => reply(new Response(JSON.stringify(savedPreview()))));
 expect(screen.queryByText('Preview status: usable')).not.toBeInTheDocument();
 view.rerender(<WritePreview owner={'owner'} draft={draft} />);
 fireEvent.click(screen.getByRole('button', {name: 'Preview exact GPS'}));
 view.rerender(<WritePreview owner={'owner'} draft={draft} hasManualConflict />);
 await act(async () => reply(new Response(JSON.stringify(savedPreview()))));
 view.rerender(<WritePreview owner={'owner'} draft={draft} />);
 expect(screen.queryByText('Preview status: usable')).not.toBeInTheDocument();
});

it('offers saved-revision reconciliation when another tab has changed the draft', async () => {
 const draft = stagedPreviewDraft(); const detail = resultDetail(); detail.entry.draftId = draft.id;
 let reads = 0;
 vi.stubGlobal('fetch', vi.fn(async (path: string) => {
  if (path.includes('/write-previews')) {return new Response(JSON.stringify({code: 'DRAFT_CONFLICT'}), {status: 409});}
  reads++;
  return new Response(JSON.stringify(reads === 1 ? draft : {...draft, revision: 2, state: 'draft', camera: {latitude: 3, longitude: 4}}), {status: 200});
 }));
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} />);
 fireEvent.click(await screen.findByRole('button', {name: 'Preview exact GPS'}));
 expect(await screen.findByText(/saved draft revision changed/)).toBeVisible();
 fireEvent.click(screen.getByRole('button', {name: 'Compare saved revision'}));
 expect(await screen.findByText('Saved elsewhere: revision 2 · Camera 3, 4')).toBeVisible();
 expect(screen.getByLabelText('Camera latitude')).toHaveValue(0);
});

it('shows the reviewed, current and proposed GPS conflict without silently acknowledging it', async () => {
 const draft = stagedPreviewDraft();
 const request = vi.fn().mockResolvedValue(new Response(JSON.stringify({code: 'IMMICH_CONFLICT', conflict: {before: {latitude: null, longitude: null}, current: {latitude: 0, longitude: 7}, proposed: draft.camera}}), {status: 409})); vi.stubGlobal('fetch', request);
 render(<WritePreview owner={'owner'} draft={draft} />);
 fireEvent.click(screen.getByRole('button', {name: 'Preview exact GPS'}));
 expect(await screen.findByText('Reviewed baseline: latitude absent · longitude absent')).toBeVisible();
 expect(screen.getByText('Current GPS: latitude 0 · longitude 7')).toBeVisible();
 expect(screen.getByText('Proposed GPS: latitude 0 · longitude 12')).toBeVisible();
 expect(screen.getByText(/Review current source and GPS.*stage/i)).toBeVisible();
 expect(screen.queryByText('Preview status: usable')).not.toBeInTheDocument();
 expect(request).toHaveBeenCalledTimes(1);
});

it('makes the displayed plan visibly expired at five minutes without renewing it', async () => {
 vi.useFakeTimers();
 try {
  const draft = stagedPreviewDraft(); const preview = savedPreview();
  const request = vi.fn().mockResolvedValue(new Response(JSON.stringify(preview), {status: 200})); vi.stubGlobal('fetch', request);
  render(<WritePreview owner={'owner'} draft={draft} />);
  fireEvent.click(screen.getByRole('button', {name: 'Preview exact GPS'}));
  await act(async () => {await vi.advanceTimersByTimeAsync(0);});
  expect(screen.getByText('Preview status: usable')).toBeVisible();
  await act(async () => {await vi.advanceTimersByTimeAsync(300_000);});
  expect(screen.getByText('Preview status: expired')).toBeVisible();
  expect(screen.getByText(`Expires: ${preview.plan.expiresAt}`)).toBeVisible();
  expect(request).toHaveBeenCalledTimes(1);
 } finally {vi.useRealTimers();}
});

it('retires a comparison when manual work overlaps or the editor becomes dirty, preserving both decisions', async () => {
 const draft = stagedPreviewDraft(); const preview = savedPreview();
 const request = vi.fn().mockImplementation(async () => new Response(JSON.stringify(preview), {status: 200})); vi.stubGlobal('fetch', request);
 const view = render(<WritePreview owner={'owner'} draft={draft} hasManualConflict />);
 expect(screen.getByRole('button', {name: 'Preview exact GPS'})).toBeDisabled();
 view.rerender(<WritePreview owner={'owner'} draft={draft} />);
 fireEvent.click(screen.getByRole('button', {name: 'Preview exact GPS'}));
 expect(await screen.findByText('Preview status: usable')).toBeVisible();
 view.rerender(<WritePreview owner={'owner'} draft={draft} hasManualConflict />);
 expect(screen.queryByText('Preview status: usable')).not.toBeInTheDocument();
 expect(screen.getByText(/Resolve the pending manual location/)).toBeVisible();
 view.rerender(<WritePreview owner={'owner'} draft={draft} />);
 expect(screen.queryByText('Preview status: usable')).not.toBeInTheDocument();
 fireEvent.click(screen.getByRole('button', {name: 'Preview exact GPS'}));
 expect(await screen.findByText('Preview status: usable')).toBeVisible();
 view.rerender(<WritePreview owner={'owner'} draft={draft} disabled />);
 expect(screen.queryByText('Preview status: usable')).not.toBeInTheDocument();
 expect(request.mock.calls.every(([path]) => String(path).includes('/ai/write-previews'))).toBe(true);
 expect(draft.camera).toEqual({latitude: 0, longitude: 12});
});

it('restores the same stored preview identifier with a read and unchanged expiry', async () => {
 const draft = stagedPreviewDraft(); const preview = savedPreview();
 sessionStorage.setItem(`ai-write-preview:owner:${draft.id}:${draft.revision}`, preview.plan.id);
 const request = vi.fn().mockResolvedValue(new Response(JSON.stringify(preview), {status: 200})); vi.stubGlobal('fetch', request);
 render(<WritePreview owner={'owner'} draft={draft} />);
 expect(await screen.findByLabelText('Before longitude')).toHaveValue('absent');
 expect(screen.getByText(`Expires: ${preview.plan.expiresAt}`)).toBeVisible();
 expect(request).toHaveBeenCalledTimes(1);
 expect(request.mock.calls[0][0]).toContain(`/ai/write-previews/${preview.plan.id}`);
 expect(request.mock.calls[0][1].method).toBe('GET');
});

it('discards an in-flight stored preview read when the private view changes', async () => {
 const draft = stagedPreviewDraft(); const preview = savedPreview();
 sessionStorage.setItem(`ai-write-preview:owner:${draft.id}:${draft.revision}`, preview.plan.id);
 let reply: (value: Response) => void = () => {};
 const request = vi.fn().mockImplementation(async () => new Promise<Response>(resolve => {reply = resolve;}));
 vi.stubGlobal('fetch', request);
 const view = render(<WritePreview owner={'owner'} draft={draft} />);
 expect(request).toHaveBeenCalledTimes(1);
 expect(request.mock.calls[0][1].method).toBe('GET');
 view.rerender(<WritePreview owner={'another-owner'} draft={{...draft, id: '66666666-6666-4666-8666-666666666666'}} />);
 await act(async () => reply(new Response(JSON.stringify(preview))));
 expect(screen.queryByLabelText('Before latitude')).not.toBeInTheDocument();
 expect(screen.queryByText(/Plan digest/)).not.toBeInTheDocument();
 expect(sessionStorage.length).toBe(0);
});
