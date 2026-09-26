import {render, screen, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import {stackIDs, stackPreview, stackReview,stagedPreviewDraft} from './testing/writePreview';
import {WritePreview} from './WritePreview';

afterEach(() => {vi.unstubAllGlobals(); sessionStorage.clear();});

it('S01 selects an exact reviewed subset of a five-photo stack', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const selected = [draft.assetId, stackIDs[0], stackIDs[2]];
 const preview = stackPreview(selected); const publish = vi.fn();
 const request = vi.fn(async (path: string, init: RequestInit) => {
  const body = JSON.parse(String(init.body));
  return new Response(JSON.stringify(path.endsWith('/stack-reviews') ? stackReview(body.targetIds?.length ? body.targetIds : undefined) : preview));
 }); vi.stubGlobal('fetch', request);
 render(<WritePreview owner={'owner'} draft={draft} onPreview={publish} />);
 await user.click(screen.getByRole('button', {name: 'Review stack members'}));
 await user.click(await screen.findByRole('checkbox', {name: `Include photo ${stackIDs[0]}`}));
 await user.click(screen.getByRole('checkbox', {name: `Include photo ${stackIDs[2]}`}));
 await user.click(screen.getByRole('button', {name: 'Read selected target baselines'}));
 for (const id of selected) {await user.click(await screen.findByRole('checkbox', {name: `Reviewed source and GPS for ${id}`}));}
 await user.click(screen.getByRole('button', {name: 'Preview exact GPS'}));
 expect(await screen.findByText('Preview status: usable')).toBeVisible();
 const comparison = screen.getByRole('region', {name: 'Approved stack target matrix'});
 for (const id of selected) {expect(within(comparison).getByText(`Photo ${id}`)).toBeVisible();}
 expect(within(comparison).queryByText(`Photo ${stackIDs[1]}`)).not.toBeInTheDocument();
 expect(publish).toHaveBeenLastCalledWith(preview);
 const post = request.mock.calls.find(([path]) => path.endsWith('/write-previews'))!;
 expect(JSON.parse(String(post[1].body))).toEqual({draftId: draft.id, draftRevision: draft.revision, stackReviewId: preview.plan.manifest.reviewId});
 expect(request.mock.calls.every(([path]) => !path.includes('/location') && !path.includes('/write-operations'))).toBe(true);
});

it('S02 retains the analyzed-photo default without implicit selection', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft();
 const preview = (await import('./testing/writePreview')).savedPreview();
 const request = vi.fn<(path: string, init: RequestInit) => Promise<Response>>(async (path: string) => new Response(JSON.stringify(path.endsWith('/stack-reviews') ? stackReview() : preview))); vi.stubGlobal('fetch', request);
 render(<WritePreview owner={'owner'} draft={draft} />);
 await user.click(screen.getByRole('button', {name: 'Review stack members'}));
 expect(await screen.findByRole('checkbox', {name: `Include photo ${draft.assetId}`})).toBeChecked();
 for (const id of stackIDs) {expect(screen.getByRole('checkbox', {name: `Include photo ${id}`})).not.toBeChecked();}
 expect(screen.queryByRole('button', {name: /select all/i})).not.toBeInTheDocument();
 await user.click(screen.getByRole('button', {name: 'Preview exact GPS'}));
 expect(await screen.findByText('Preview status: usable')).toBeVisible();
 expect(screen.queryByRole('region', {name: 'Approved stack target matrix'})).not.toBeInTheDocument();
 expect(JSON.parse(String(request.mock.calls.find(([path]) => path.endsWith('/write-previews'))?.[1]?.body))).toEqual({draftId: draft.id, draftRevision: draft.revision});
});

it('S20 prevents stack expansion without selected valid GPS', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft();
 const request = vi.fn(async () => new Response(JSON.stringify(stackReview()))); vi.stubGlobal('fetch', request);
 const view = render(<WritePreview owner={'owner'} draft={{...draft, fields: ['description'], camera: null}} />);
 expect(screen.getByRole('button', {name: 'Review stack members'})).toBeDisabled();
 view.rerender(<WritePreview owner={'owner'} draft={{...draft, camera: {latitude: 91, longitude: 0}}} />);
 expect(screen.getByRole('button', {name: 'Review stack members'})).toBeDisabled();
 view.rerender(<WritePreview owner={'owner'} draft={draft} />);
 await user.click(screen.getByRole('button', {name: 'Review stack members'}));
 expect(await screen.findByRole('checkbox', {name: `Include photo ${stackIDs[0]}`})).toBeEnabled();
 expect(request).toHaveBeenCalledTimes(1);
});

it('S17 P14 exposes the complete target field matrix by keyboard with failed tiles in a narrow viewport', async () => {
 const {act, waitFor} = await import('@testing-library/react'); const mapBase = await import('@/features/map');
 const {DraftReview} = await import('./DraftReview'); const {parseReview} = await import('./reviewParser');
 const {resultDetail} = await import('./testing/resultDetail'); const {descriptionDraft} = await import('./testing/draft');
 const user = userEvent.setup(); const map = vi.spyOn(mapBase, 'createBaseMap'); vi.stubGlobal('innerWidth', 360);
 const draft = {...stagedPreviewDraft(), descriptions: descriptionDraft().descriptions, primaryLanguage: 'uk', descriptionPolicy: 'replace' as const, fields: ['gps', 'description'] as ('gps' | 'description')[]};
 draft.baseline.description = descriptionDraft().baseline.description;
 const preview = stackPreview(); const description = {before: draft.baseline.description!, intended: draft.descriptions[0].text!, language: 'uk', policy: 'replace'};
 const plan = {...preview.plan, fields: draft.fields, description, manifest: {...preview.plan.manifest, targets: preview.plan.manifest.targets.map(target => target.assetId === draft.assetId ? {...target, fields: draft.fields, description} : target)}};
 const detail = resultDetail(); detail.entry.draftId = draft.id;
 const request = vi.fn(async (path: string, init: RequestInit) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : path.endsWith('/stack-reviews') ? stackReview(JSON.parse(String(init.body)).targetIds) : path.endsWith('/write-previews') ? {...preview, plan} : draft))); vi.stubGlobal('fetch', request);
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} />);
 await waitFor(() => expect(map).toHaveBeenCalled()); act(() => {map.mock.results[0].value.tiles.fire('tileerror');});
 expect(screen.getByText(/Draft map tiles unavailable/)).toBeVisible();
 (await screen.findByRole('button', {name: 'Review stack members'})).focus(); await user.keyboard('{Enter}');
 (await screen.findByRole('checkbox', {name: `Include photo ${stackIDs[0]}`})).focus(); await user.keyboard(' ');
 screen.getByRole('button', {name: 'Read selected target baselines'}).focus(); await user.keyboard('{Enter}');
 for (const id of [draft.assetId, stackIDs[0]]) {(await screen.findByRole('checkbox', {name: `Reviewed source and GPS for ${id}`})).focus(); await user.keyboard(' ');}
 screen.getByRole('button', {name: 'Preview selected fields'}).focus(); await user.keyboard('{Enter}');
 expect(await screen.findByRole('button', {name: 'Confirm selected fields write'})).toBeEnabled();
 expect(screen.queryByText(/Single photo ·/)).not.toBeInTheDocument();
 const primary = within(screen.getByRole('region', {name: `Approved photo ${draft.assetId}`}));
 const sibling = within(screen.getByRole('region', {name: `Approved photo ${stackIDs[0]}`}));
 expect(primary.getByLabelText('Before latitude')).toHaveValue('0'); expect(primary.getByLabelText('Before longitude')).toHaveValue('absent');
 expect(sibling.getByLabelText('Before latitude')).toHaveValue('absent'); expect(sibling.getByLabelText('Before longitude')).toHaveValue('7');
 expect(sibling.getByLabelText('Proposed latitude')).toHaveValue('0'); expect(sibling.getByLabelText('Proposed longitude')).toHaveValue('12');
 expect(primary.getByLabelText('Proposed description').textContent).toBe(description.intended); expect(sibling.queryByLabelText('Proposed description')).not.toBeInTheDocument();
 expect(screen.getByText('Description language: uk · Policy: replace')).toBeVisible(); expect(screen.getByText(/Draft revision 1/)).toBeVisible(); expect(screen.getByText(/^Expires:/)).toBeVisible();
 expect(request.mock.calls.some(([path]) => path.endsWith('/write-operations'))).toBe(false);
 map.mockRestore();
});

it('bounds individual selection at fifty without silently dropping chosen photos', async () => {
 const {fireEvent} = await import('@testing-library/react'); const draft = stagedPreviewDraft();
 const candidates = [draft.assetId, ...Array.from({length: 50}, (_, index) => `00000000-0000-4000-8000-${String(index).padStart(12, '0')}`)];
 vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({...stackReview(), candidates}))));
 render(<WritePreview owner={'owner'} draft={draft} />);
 fireEvent.click(screen.getByRole('button', {name: 'Review stack members'}));
 await screen.findByRole('checkbox', {name: `Include photo ${candidates[1]}`});
 for (const id of candidates.slice(1, 50)) {fireEvent.click(screen.getByRole('checkbox', {name: `Include photo ${id}`}));}
 expect(screen.getByRole('checkbox', {name: `Include photo ${candidates[50]}`})).toBeDisabled();
 expect(screen.getAllByRole('checkbox').filter(box => (box as HTMLInputElement).checked)).toHaveLength(50);
 expect(screen.getByText('50 selected · Maximum 50 including the analyzed photo.')).toBeVisible();
});

it('expires reviewed target acknowledgements without renewing authority', async () => {
 vi.useFakeTimers();
 try {
  const {act, fireEvent} = await import('@testing-library/react'); const draft = stagedPreviewDraft();
  const request = vi.fn(async (_path: string, init: RequestInit) => new Response(JSON.stringify(stackReview(JSON.parse(String(init.body)).targetIds)))); vi.stubGlobal('fetch', request);
  render(<WritePreview owner={'owner'} draft={draft} />);
  fireEvent.click(screen.getByRole('button', {name: 'Review stack members'})); await act(async () => {await vi.advanceTimersByTimeAsync(0);});
  fireEvent.click(screen.getByRole('checkbox', {name: `Include photo ${stackIDs[0]}`}));
  fireEvent.click(screen.getByRole('button', {name: 'Read selected target baselines'})); await act(async () => {await vi.advanceTimersByTimeAsync(0);});
  for (const id of [draft.assetId, stackIDs[0]]) {fireEvent.click(screen.getByRole('checkbox', {name: `Reviewed source and GPS for ${id}`}));}
  expect(screen.getByRole('button', {name: 'Preview exact GPS'})).toBeEnabled();
  await act(async () => {await vi.advanceTimersByTimeAsync(300_000);});
  expect(screen.getByRole('button', {name: 'Preview exact GPS'})).toBeDisabled();
  expect(screen.getByText('Target review expired. Read and acknowledge the selected baselines again.')).toBeVisible();
  expect(request).toHaveBeenCalledTimes(2);
 } finally {vi.useRealTimers();}
});

it('fences a pending target observation when manual overlap appears and requires a fresh read', async () => {
 const {act} = await import('@testing-library/react'); const user = userEvent.setup(); const draft = stagedPreviewDraft();
 let release: (value: Response) => void = () => undefined; let reads = 0;
 const request = vi.fn(async () => ++reads === 1 ? new Response(JSON.stringify(stackReview())) : new Promise<Response>(resolve => {release = resolve;})); vi.stubGlobal('fetch', request);
 const view = render(<WritePreview owner={'owner'} draft={draft} />);
 await user.click(screen.getByRole('button', {name: 'Review stack members'})); await user.click(await screen.findByRole('checkbox', {name: `Include photo ${stackIDs[0]}`}));
 await user.click(screen.getByRole('button', {name: 'Read selected target baselines'}));
 view.rerender(<WritePreview owner={'owner'} draft={draft} manualPendingIDs={[stackIDs[0]]} />);
 await act(async () => release(new Response(JSON.stringify(stackReview([draft.assetId, stackIDs[0]])))));
 view.rerender(<WritePreview owner={'owner'} draft={draft} manualPendingIDs={[]} />);
 expect(screen.queryByRole('checkbox', {name: `Reviewed source and GPS for ${stackIDs[0]}`})).not.toBeInTheDocument();
 expect(screen.getByRole('button', {name: 'Preview exact GPS'})).toBeDisabled();
 expect(screen.getByRole('button', {name: 'Read selected target baselines'})).toBeEnabled();
});

it('preserves the deliberately selected subset when refreshing member review', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft();
 const request = vi.fn(async (_path: string, init: RequestInit) => new Response(JSON.stringify(stackReview(JSON.parse(String(init.body)).targetIds)))); vi.stubGlobal('fetch', request);
 render(<WritePreview owner={'owner'} draft={draft} />);
 await user.click(screen.getByRole('button', {name: 'Review stack members'}));
 await user.click(await screen.findByRole('checkbox', {name: `Include photo ${stackIDs[0]}`}));
 await user.click(screen.getByRole('button', {name: 'Review stack members'}));
 expect(JSON.parse(String(request.mock.calls[1][1].body)).targetIds).toEqual([draft.assetId, stackIDs[0]].sort());
 expect(screen.getByRole('checkbox', {name: `Include photo ${stackIDs[0]}`})).toBeChecked();
 expect(screen.getByRole('button', {name: 'Preview exact GPS'})).toBeDisabled();
});

it('retains an immutable plan until its own expiry when the earlier target review expires', async () => {
 vi.useFakeTimers();
 try {
  const {act, fireEvent} = await import('@testing-library/react'); const draft = stagedPreviewDraft(); const ids = [draft.assetId, stackIDs[0]];
  const review = stackReview(ids); const published = vi.fn();
  vi.stubGlobal('fetch', vi.fn(async (path: string, init: RequestInit) => new Response(JSON.stringify(path.endsWith('/stack-reviews') ? stackReview(JSON.parse(String(init.body)).targetIds) : stackPreview(ids)))));
  render(<WritePreview owner={'owner'} draft={draft} onPreview={published} />);
  fireEvent.click(screen.getByRole('button', {name: 'Review stack members'})); await act(async () => {await vi.advanceTimersByTimeAsync(0);});
  fireEvent.click(screen.getByRole('checkbox', {name: `Include photo ${stackIDs[0]}`})); fireEvent.click(screen.getByRole('button', {name: 'Read selected target baselines'})); await act(async () => {await vi.advanceTimersByTimeAsync(0);});
  for (const id of ids) {fireEvent.click(screen.getByRole('checkbox', {name: `Reviewed source and GPS for ${id}`}));}
  await act(async () => {await vi.advanceTimersByTimeAsync(60_000);});
  fireEvent.click(screen.getByRole('button', {name: 'Preview exact GPS'})); await act(async () => {await vi.advanceTimersByTimeAsync(0);});
  expect(screen.getByText('Preview status: usable')).toBeVisible();
  await act(async () => {await vi.advanceTimersByTimeAsync(Date.parse(review.expiresAt) - Date.now());});
  expect(screen.getByText('Preview status: usable')).toBeVisible();
  expect(published.mock.calls.at(-1)?.[0]?.plan.version).toBe('stack-preview-v3');
  expect(screen.getByRole('button', {name: 'Preview exact GPS'})).toBeDisabled();
 } finally {vi.useRealTimers();}
});
