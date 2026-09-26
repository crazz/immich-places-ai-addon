import {act, render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import * as mapBase from '@/features/map';

import {DraftReview} from './DraftReview';
import {parseReview} from './reviewParser';
import {mirrorPreview, mirroredDraft} from './testing/mirror';
import {resultDetail} from './testing/resultDetail';

afterEach(() => {vi.unstubAllGlobals(); vi.restoreAllMocks(); sessionStorage.clear();});

it('M19 P14 exposes exact export and reader disclosure by keyboard with unavailable map tiles', async () => {
 const user = userEvent.setup(); const draft = mirroredDraft(); const preview = mirrorPreview(draft); const detail = resultDetail(); detail.entry.draftId = draft.id;
 const initial = {...draft, mirror: undefined}; const map = vi.spyOn(mapBase, 'createBaseMap'); vi.stubGlobal('innerWidth', 360);
 const request = vi.fn(async (path: string, init: RequestInit) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : path.endsWith('/write-previews') ? preview : init.method === 'PATCH' ? draft : initial))); vi.stubGlobal('fetch', request);
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} />);
 await waitFor(() => expect(map).toHaveBeenCalled()); act(() => {map.mock.results[0].value.tiles.fire('tileerror');});
 const mirror = await screen.findByLabelText('Mirror optional metadata for this photo'); mirror.focus(); await user.keyboard(' ');
 screen.getByLabelText('Mirror direction').focus(); await user.keyboard(' '); screen.getByLabelText('Mirror description uk').focus(); await user.keyboard(' ');
 screen.getByRole('button', {name: 'Stage GPS draft'}).focus(); await user.keyboard('{Enter}');
 const consent = await screen.findByLabelText('I understand authorized asset readers may see this exported information');
 expect(screen.getByRole('button', {name: 'Preview exact GPS'})).toBeDisabled(); consent.focus(); await user.keyboard(' ');
 screen.getByRole('button', {name: 'Preview exact GPS'}).focus(); await user.keyboard('{Enter}');
 expect(await screen.findByText('Preview status: usable')).toBeVisible();
 const exact = screen.getByText('Exact metadata before and after'); exact.focus(); await user.keyboard('{Enter}');
 expect(JSON.parse(screen.getByLabelText('Before metadata').textContent!)).toEqual({present: false});
 expect(JSON.parse(screen.getByLabelText('Proposed metadata').textContent!)).toEqual(preview.plan.mirror.value);
 expect(screen.getByText('Mirrored direction: unknown')).toBeVisible(); expect(screen.getByText(`Mirror target: ${draft.assetId}`)).toBeVisible();
 expect(screen.getByText('Step order: verify this photo’s standard fields, then its optional metadata.')).toBeVisible();
 expect(await screen.findByRole('button', {name: 'Confirm GPS write'})).toBeEnabled();
 const post = request.mock.calls.find(([path]) => path.endsWith('/write-previews'))!;
 expect(JSON.parse(String(post[1].body))).toEqual({draftId: draft.id, draftRevision: draft.revision, mirrorDisclosure: 'asset-readers-v1'});
 expect(request.mock.calls.some(([path]) => path.endsWith('/write-operations'))).toBe(false);
});

it('M01 retains local direction and translations on unsupported mirroring and permits explicit standard-only review', async () => {
 const user = userEvent.setup(); const draft = mirroredDraft(); const detail = resultDetail(); detail.entry.draftId = draft.id; let isStandard = false;
 const {savedPreview} = await import('./testing/writePreview'); const standard = {...draft, revision: 2, mirror: undefined}; const preview = savedPreview(); preview.plan.draftRevision = 2;
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (path.endsWith('/write-previews')) {return isStandard ? new Response(JSON.stringify(preview)) : new Response(JSON.stringify({code: 'METADATA_UNSUPPORTED'}), {status: 409});}
  if (init.method === 'PATCH') {isStandard = true; return new Response(JSON.stringify(standard));}
  return new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : draft));
 }); vi.stubGlobal('fetch', request);
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} />);
 await user.click(await screen.findByLabelText('I understand authorized asset readers may see this exported information')); await user.click(screen.getByRole('button', {name: 'Preview exact GPS'}));
 expect(await screen.findByText(/Metadata mirroring is unsupported or unverified/)).toBeVisible();
 expect(screen.getByLabelText('Description uk')).toHaveValue(draft.descriptions[0].text); expect(screen.getByLabelText('Heading degrees')).toHaveValue(null); expect(screen.getByLabelText('Mirror optional metadata for this photo')).toBeChecked();
 expect(request.mock.calls.filter(([path]) => path.endsWith('/write-previews'))).toHaveLength(1);
 await user.click(screen.getByLabelText('Mirror optional metadata for this photo')); await user.click(screen.getByRole('button', {name: 'Stage GPS draft'}));
 await waitFor(() => expect(screen.getByText('Saved staged · Revision 2')).toBeVisible());
 await user.click(screen.getByRole('button', {name: 'Preview exact GPS'})); expect(await screen.findByText('Preview status: usable')).toBeVisible();
 const posts = request.mock.calls.filter(([path]) => path.endsWith('/write-previews'));
 expect(JSON.parse(String(posts[1][1].body))).toEqual({draftId: draft.id, draftRevision: 2});
 expect(JSON.parse(String(request.mock.calls.find(([, init]) => init.method === 'PATCH')![1].body)).mirror).toBeNull();
 expect(request.mock.calls.some(([path]) => path.endsWith('/write-operations'))).toBe(false);
});

it('M06 P16 invalidates disclosure and fences the old mirror comparison after an edited revision', async () => {
 const {WritePreview} = await import('./WritePreview'); const user = userEvent.setup(); const draft = mirroredDraft(); const preview = mirrorPreview(draft); const publish = vi.fn();
 let release: (response: Response) => void = () => undefined; const request = vi.fn(async () => new Promise<Response>(resolve => {release = resolve;})); vi.stubGlobal('fetch', request);
 const view = render(<WritePreview owner={'owner'} draft={draft} onPreview={publish} />);
 await user.click(screen.getByLabelText('I understand authorized asset readers may see this exported information')); await user.click(screen.getByRole('button', {name: 'Preview exact GPS'}));
 view.rerender(<WritePreview owner={'owner'} draft={{...draft, revision: 2, descriptions: [{...draft.descriptions[0], text: 'New reviewed text'}]}} onPreview={publish} />);
 expect(screen.getByLabelText('I understand authorized asset readers may see this exported information')).not.toBeChecked();
 await act(async () => release(new Response(JSON.stringify(preview))));
 expect(screen.queryByLabelText('Proposed metadata')).not.toBeInTheDocument(); expect(publish).toHaveBeenLastCalledWith(null);
 view.rerender(<WritePreview owner={'other'} draft={{...draft, id: '55555555-5555-4555-8555-555555555555'}} onPreview={publish} />);
 expect(screen.queryByText('Preview status: usable')).not.toBeInTheDocument();
 expect(request).toHaveBeenCalledTimes(1);
});

it('withdraws displayed mirror authority when its disclosure acknowledgement is cleared', async () => {
 const {WritePreview} = await import('./WritePreview'); const user = userEvent.setup(); const draft = mirroredDraft(); const preview = mirrorPreview(draft); const publish = vi.fn();
 vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify(preview))));
 render(<WritePreview owner={'owner'} draft={draft} onPreview={publish} />);
 await user.click(screen.getByLabelText('I understand authorized asset readers may see this exported information')); await user.click(screen.getByRole('button', {name: 'Preview exact GPS'}));
 expect(await screen.findByText('Preview status: usable')).toBeVisible();
 await user.click(screen.getByLabelText('I understand authorized asset readers may see this exported information'));
 expect(publish).toHaveBeenLastCalledWith(null); expect(screen.queryByText('Preview status: usable')).not.toBeInTheDocument();
 expect(screen.getByRole('button', {name: 'Preview exact GPS'})).toBeDisabled();
});

it('P15 preserves exact stack scope and requires explicit manual overlap resolution for combined mirroring', async () => {
 const {stackPreview, stackIDs} = await import('./testing/writePreview'); const draft = mirroredDraft(); const original = mirrorPreview(draft); const stack = stackPreview();
 const preview = {...original, plan: {...original.plan, manifest: stack.plan.manifest}}; const detail = resultDetail(); detail.entry.draftId = draft.id;
 sessionStorage.setItem(`ai-write-preview:owner:${draft.id}:${draft.revision}`, preview.plan.id);
 const request = vi.fn(async (path: string) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : path.includes('/write-previews/') ? preview : draft))); vi.stubGlobal('fetch', request);
 const unrelated = stackIDs[2]; const pending = [unrelated];
 const view = render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} manualPendingIDs={pending} />);
 expect(await screen.findByRole('region', {name: `Approved photo ${stackIDs[0]}`})).toBeVisible();
 expect(await screen.findByRole('button', {name: 'Confirm GPS write'})).toBeEnabled();
 view.rerender(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} manualPendingIDs={[...pending, stackIDs[0]]} />);
 await waitFor(() => expect(screen.queryByRole('button', {name: 'Confirm GPS write'})).not.toBeInTheDocument());
 expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeDisabled(); expect(screen.getByText(`Resolve pending manual GPS for ${stackIDs[0]} by explicitly saving or discarding that manual change.`)).toBeVisible();
 expect(pending).toEqual([unrelated]); expect(request.mock.calls.some(([path]) => path.endsWith('/write-operations') || path.endsWith('/stack-reviews'))).toBe(false);
});

it('confirms description and metadata only while preserving pending manual GPS and gallery membership', async () => {
 const {descriptionDraft} = await import('./testing/draft'); const {descriptionPreview} = await import('./testing/writePreview'); const {mirrorOperation} = await import('./testing/mirror'); const {within} = await import('@testing-library/react');
 const user = userEvent.setup(); const draft = {...mirroredDraft(), ...descriptionDraft(), state: 'staged' as const, fields: ['description' as const], primaryLanguage: 'uk', descriptionPolicy: 'replace' as const, headingStale: false};
 const standard = descriptionPreview(draft); const initial = mirrorPreview();
 const preview = {...initial, plan: {...standard.plan, version: 'mirror-preview-v4' as const, comparisonPolicy: 'standard-then-metadata-v4' as const, mirror: initial.plan.mirror, manifest: {...initial.plan.manifest, targets: [{...initial.plan.manifest.targets[0], fields: draft.fields, before: {latitude: null, longitude: null}, intended: {latitude: 0, longitude: 0}, description: standard.plan.description}]}}};
 const op = {...mirrorOperation(), plan: preview.plan, refreshed: false, targets: [{...mirrorOperation().targets[0], refreshed: false, observed: null, fields: [{field: 'description', status: 'verified', wasVerified: true, description: {presence: 'value', value: standard.plan.description.intended}}]}]};
 const detail = resultDetail(); detail.entry.draftId = draft.id; const refresh = vi.fn(); const pending = [draft.assetId];
 const request = vi.fn(async (path: string) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : path.endsWith('/write-previews') ? preview : path.endsWith('/write-operations') ? op : draft))); vi.stubGlobal('fetch', request);
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} hasManualConflict manualPendingIDs={pending} onVerified={refresh} />);
 await user.click(await screen.findByLabelText('I understand authorized asset readers may see this exported information')); await user.click(screen.getByRole('button', {name: 'Preview selected fields'}));
 const target = within(await screen.findByRole('region', {name: `Approved photo ${draft.assetId}`}));
 expect(target.queryByLabelText('Proposed latitude')).not.toBeInTheDocument(); expect(target.getByLabelText('Proposed description').textContent).toBe(draft.descriptions[0].text);
 expect(screen.getByText('Confirming authorizes the exact standard-field matrix and analyzed-photo metadata export shown above. Standard fields run before metadata. Manual pending choices are preserved.')).toBeVisible();
 await user.click(screen.getByRole('button', {name: 'Confirm selected fields write'})); expect(await screen.findByText('Metadata step: retryable')).toBeVisible();
 const writes = request.mock.calls.filter(([path]) => path.endsWith('/write-operations')); expect(writes).toHaveLength(1); expect(pending).toEqual([draft.assetId]); expect(refresh).not.toHaveBeenCalled();
});

it('explains typed metadata preview failures without discarding selected content or silently falling back', async () => {
 const {WritePreview} = await import('./WritePreview'); const user = userEvent.setup(); const draft = mirroredDraft();
 const cases = [
  ['INVALID_MIRROR', 'Selected metadata needs current review. Review or deselect stale contents, save and stage a new revision.'],
  ['MIRROR_TOO_LARGE', 'Selected metadata exceeds the supported size. Choose fewer languages or shorten reviewed text, then stage a new revision.'],
  ['METADATA_UNAVAILABLE', 'Current metadata is unavailable. Your draft is preserved; retry the comparison when access returns.'],
  ['METADATA_CONFLICT', 'Owned metadata changed. Create a fresh comparison and review the full existing export before confirming.'],
 ];
 const request = vi.fn<(path: string) => Promise<Response>>(async () => new Response(JSON.stringify({code: cases.shift()![0]}), {status: 409})); vi.stubGlobal('fetch', request);
 render(<WritePreview owner={'owner'} draft={draft} />); await user.click(screen.getByLabelText('I understand authorized asset readers may see this exported information'));
 while (cases.length) {const message = cases[0][1]; await user.click(screen.getByRole('button', {name: 'Preview exact GPS'})); expect(await screen.findByText(message)).toBeVisible(); expect(screen.queryByText('Preview status: usable')).not.toBeInTheDocument();}
 expect(request).toHaveBeenCalledTimes(4); expect(request.mock.calls.every(([path]) => String(path).endsWith('/write-previews'))).toBe(true); expect(draft.mirror.languages).toEqual(['uk']);
});

it('describes optional metadata scope without claiming all metadata is preserved', async () => {
 const {WritePreview} = await import('./WritePreview'); const draft = mirroredDraft();
 render(<WritePreview owner={'owner'} draft={draft} />);
 expect(screen.getByText('Selected standard fields and the analyzed photo’s optional metadata will be compared independently.')).toBeVisible();
 expect(screen.queryByText('Single photo · GPS only. Other metadata is preserved.')).not.toBeInTheDocument();
});
