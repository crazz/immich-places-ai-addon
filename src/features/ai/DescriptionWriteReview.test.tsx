import {act, render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import * as mapBase from '@/features/map';

import {DraftReview} from './DraftReview';
import {parseReview} from './reviewParser';
import {descriptionDraft} from './testing/draft';
import {resultDetail} from './testing/resultDetail';
import {descriptionPreview} from './testing/writePreview';


afterEach(() => {vi.restoreAllMocks(); vi.unstubAllGlobals(); sessionStorage.clear();});

it('C01 stages and confirms a scene-only description without camera coordinates', async () => {
 const user = userEvent.setup(); let draft = descriptionDraft(); const detail = resultDetail(); detail.entry.draftId = draft.id;
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (path.includes('?draftId=')) {return new Response(JSON.stringify({items: [], nextCursor: ''}));}
  if (path.includes('/write-previews')) {return new Response(JSON.stringify(descriptionPreview(draft)));}
  if (path.endsWith('/write-operations')) {const preview = descriptionPreview(draft); return new Response(JSON.stringify({id: detail.entry.id, plan: preview.plan, digest: preview.digest, status: 'succeeded', code: 'STANDARD_VERIFIED', approvedAt: new Date().toISOString(), attempts: 1, generation: 1, observed: null, verified: true, refreshed: false, noop: false, settled: true, events: [], fields: [{field: 'description', status: 'verified', wasVerified: true, description: {presence: 'value', value: draft.descriptions[0].text}}]}));}
  if (init.method === 'PATCH') {draft = {...draft, ...JSON.parse(String(init.body)), revision: draft.revision + 1};}
  return new Response(JSON.stringify(draft));
 }); vi.stubGlobal('fetch', request);
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} />);
 await user.selectOptions(await screen.findByLabelText('Primary description language'), 'uk');
 await user.selectOptions(screen.getByLabelText('Description policy'), 'replace');
 await user.click(screen.getByLabelText('Select description for this photo'));
 await user.click(screen.getByRole('button', {name: 'Stage selected fields'}));
 expect(await screen.findByText('Saved staged · Revision 2')).toBeVisible();
 await user.click(screen.getByRole('button', {name: 'Preview selected fields'}));
 expect(await screen.findByLabelText('Before description')).toHaveTextContent('Before');
 expect(screen.getByLabelText('Before description').textContent).toBe('  Before\n前  ');
 expect(screen.getByLabelText('Proposed description').textContent).toBe('  Місто\n河  ');
 expect(screen.queryByLabelText('Proposed latitude')).not.toBeInTheDocument();
 await user.click(screen.getByRole('button', {name: 'Confirm selected fields write'}));
 expect(await screen.findByText('Description verified; no GPS catalog refresh is needed.')).toBeVisible();
 expect(screen.getByText('Selected fields operation: succeeded')).toBeVisible();
 await waitFor(() => expect(request.mock.calls.filter(([path, init]) => path.endsWith('/write-operations') && init.method === 'POST')).toHaveLength(1));
 expect(draft.camera).toBeNull(); expect(draft.fields).toEqual(['description']);
 expect(request.mock.calls.some(([path]) => path.includes('/location'))).toBe(false);
});

it('C04 displays exact absent, empty, Unicode and multiline baselines through preview reload', async () => {
 const user = userEvent.setup();
 for (const before of [{presence: 'absent' as const, value: ''}, {presence: 'null' as const, value: ''}, {presence: 'value' as const, value: ''}, {presence: 'value' as const, value: '  e\u0301\nМісто\r\n河  '}]) {
  const original = descriptionDraft(); const draft = {...original, state: 'staged' as const, fields: ['description' as const], primaryLanguage: 'uk', descriptionPolicy: 'replace' as const, baseline: {...original.baseline, description: before}};
  const detail = resultDetail(); detail.entry.draftId = draft.id; const preview = descriptionPreview(draft);
  const request = vi.fn(async (path: string) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : path.includes('/write-previews') ? preview : draft)));
  vi.stubGlobal('fetch', request);
  const view = render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} />);
  expect((await screen.findByLabelText('Reviewed description')).textContent).toBe(before.value);
  expect(screen.getByText(`Reviewed description presence: ${before.presence}`)).toBeVisible();
  await user.click(screen.getByRole('button', {name: 'Preview selected fields'}));
  expect((await screen.findByLabelText('Before description')).textContent).toBe(before.value);
  view.unmount();
  sessionStorage.setItem(`ai-write-preview:owner:${draft.id}:${draft.revision}`, preview.plan.id);
  const restored = render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} />);
  expect((await screen.findByLabelText('Before description')).textContent).toBe(before.value);
  expect(screen.getByText(`Expires: ${preview.plan.expiresAt}`)).toBeVisible();
  expect(request.mock.calls.filter(([path]) => path.includes('/write-previews/'))).toHaveLength(1);
  expect(request.mock.calls.some(([path]) => path.endsWith('/write-operations'))).toBe(false);
  restored.unmount(); sessionStorage.clear();
 }
});

it('C05 retires a preview when another client changes the selected description and requires renewed baseline review', async () => {
 const user = userEvent.setup(); const draft = {...descriptionDraft(), state: 'staged' as const, fields: ['description' as const], primaryLanguage: 'uk', descriptionPolicy: 'replace' as const};
 const detail = resultDetail(); detail.entry.draftId = draft.id; let previews = 0;
 const request = vi.fn(async (path: string) => {
  if (path.includes('/write-previews') && ++previews > 1) {return new Response(JSON.stringify({code: 'DESCRIPTION_CONFLICT'}), {status: 409});}
  return new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : path.includes('/write-previews') ? descriptionPreview(draft) : draft));
 }); vi.stubGlobal('fetch', request);
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} />);
 await user.click(await screen.findByRole('button', {name: 'Preview selected fields'}));
 expect(await screen.findByRole('button', {name: 'Confirm selected fields write'})).toBeEnabled();
 await user.click(screen.getByRole('button', {name: 'Preview selected fields'}));
 expect(await screen.findByText(/Description changed.*Review current source and selected fields/)).toBeVisible();
 expect(screen.queryByRole('button', {name: 'Confirm selected fields write'})).not.toBeInTheDocument();
 expect(screen.getByLabelText('Description uk')).toHaveValue(draft.descriptions[0].text);
 expect(request.mock.calls.some(([path]) => path.endsWith('/write-operations'))).toBe(false);
});

it('C18 exposes full managed-append confirmation by keyboard with unavailable map and narrow viewport', async () => {
 const user = userEvent.setup(); const map = vi.spyOn(mapBase, 'createBaseMap'); vi.stubGlobal('innerWidth', 360);
 let draft = descriptionDraft(); const detail = resultDetail(); detail.entry.draftId = draft.id;
 draft.baseline.description = {presence: 'value', value: `  ${'長'.repeat(6000)}\n  User text  `};
 const preview = (): ReturnType<typeof descriptionPreview> => {
  const value = descriptionPreview(draft); const id = '55555555-5555-4555-8555-555555555555';
  const block = `[[Immich Places AI v1:${id}]]\nLanguage: uk\n${draft.descriptions[0].text}\n[[/Immich Places AI v1:${id}]]`;
  return {...value, plan: {...value.plan, description: {...value.plan.description, policy: 'managed_append', intended: `${draft.baseline.description!.value}\n\n${block}`, lineage: {id, block, hash: 'd'.repeat(64)}}}};
 };
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (init.method === 'PATCH') {draft = {...draft, ...JSON.parse(String(init.body)), revision: draft.revision + 1};}
  return new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : path.includes('/write-previews') ? preview() : draft));
 }); vi.stubGlobal('fetch', request);
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} />);
 await waitFor(() => expect(map).toHaveBeenCalled());
 act(() => {map.mock.results[0].value.tiles.fire('tileerror');});
 expect(screen.getByText(/Draft map tiles unavailable/)).toBeVisible();
 await user.selectOptions(screen.getByLabelText('Primary description language'), 'uk');
 await user.selectOptions(screen.getByLabelText('Description policy'), 'managed_append');
 screen.getByLabelText('Select description for this photo').focus(); await user.keyboard(' ');
 screen.getByRole('button', {name: 'Stage selected fields'}).focus(); await user.keyboard('{Enter}');
 await screen.findByText('Saved staged · Revision 2');
 screen.getByRole('button', {name: 'Preview selected fields'}).focus(); await user.keyboard('{Enter}');
 const confirmation = await screen.findByRole('button', {name: 'Confirm selected fields write'});
 expect(confirmation).toBeEnabled();
 const displayed = screen.getByLabelText('Proposed description'); displayed.focus(); expect(displayed).toHaveFocus();
 expect(displayed.textContent).toBe(preview().plan.description.intended);
 expect(screen.getByLabelText('Before description').textContent).toBe(draft.baseline.description!.value);
 expect(screen.getByText('Selected fields: description')).toBeVisible();
 expect(screen.getByText('Description language: uk · Policy: managed_append')).toBeVisible();
 expect(screen.getByText(/Draft revision 2/)).toBeVisible(); expect(screen.getByText(/^Expires:/)).toBeVisible();
 expect(request.mock.calls.some(([path]) => path.endsWith('/write-operations'))).toBe(false);
});

it('C19 completes a description-only operation while preserving pending manual GPS and gallery membership', async () => {
 const user = userEvent.setup(); const draft = {...descriptionDraft(), state: 'staged' as const, fields: ['description' as const], primaryLanguage: 'uk', descriptionPolicy: 'replace' as const};
 const detail = resultDetail(); detail.entry.draftId = draft.id; const preview = descriptionPreview(draft); const refresh = vi.fn();
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (path.endsWith('/write-operations') && init.method === 'POST') {return new Response(JSON.stringify({id: detail.entry.id, plan: preview.plan, digest: preview.digest, status: 'succeeded', code: 'STANDARD_VERIFIED', approvedAt: new Date().toISOString(), attempts: 1, generation: 1, observed: null, verified: true, refreshed: false, noop: false, settled: true, events: [], fields: [{field: 'description', status: 'verified', wasVerified: true, description: {presence: 'value', value: draft.descriptions[0].text}}]}));}
  return new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : path.includes('/write-previews') ? preview : draft));
 }); vi.stubGlobal('fetch', request);
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} hasManualConflict onVerified={refresh} />);
 const action = await screen.findByRole('button', {name: 'Preview selected fields'}); expect(action).toBeEnabled(); await user.click(action);
 await user.click(await screen.findByRole('button', {name: 'Confirm selected fields write'}));
 expect(await screen.findByText('Description verified; no GPS catalog refresh is needed.')).toBeVisible();
 expect(screen.queryByText(/Resolve the pending manual location/)).not.toBeInTheDocument();
 expect(refresh).not.toHaveBeenCalled(); expect(draft.camera).toBeNull();
 expect(request.mock.calls.filter(([, init]) => init.method === 'POST')).toHaveLength(2);
 expect(request.mock.calls.some(([path]) => path.includes('/location'))).toBe(false);
});

it('invalidates description preview immediately on policy or language edits and saves a new unstaged revision', async () => {
 const user = userEvent.setup(); let draft = {...descriptionDraft(), state: 'staged' as 'draft' | 'staged', fields: ['description' as const], primaryLanguage: 'uk', descriptionPolicy: 'replace' as 'replace' | 'managed_append'};
 draft.descriptions.push({...draft.descriptions[0], language: 'en', text: 'Scene'});
 const detail = resultDetail(); detail.entry.draftId = draft.id;
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (init.method === 'PATCH') {draft = {...draft, ...JSON.parse(String(init.body)), revision: draft.revision + 1, state: 'draft'};}
  return new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : path.includes('/write-previews') ? descriptionPreview(draft) : draft));
 }); vi.stubGlobal('fetch', request);
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} />);
 await user.click(await screen.findByRole('button', {name: 'Preview selected fields'}));
 await screen.findByRole('button', {name: 'Confirm selected fields write'});
 await user.selectOptions(screen.getByLabelText('Description policy'), 'managed_append');
 expect(screen.queryByRole('button', {name: 'Confirm selected fields write'})).not.toBeInTheDocument();
 expect(screen.queryByLabelText('Before description')).not.toBeInTheDocument();
 await user.selectOptions(screen.getByLabelText('Primary description language'), 'en');
 await user.click(screen.getByRole('button', {name: 'Save draft'}));
 expect(await screen.findByText('Saved draft · Revision 2')).toBeVisible();
 expect(screen.getByRole('button', {name: 'Preview selected fields'})).toBeDisabled();
 const patch = request.mock.calls.find(([, init]) => init.method === 'PATCH')![1];
 expect(new Headers(patch.headers).get('If-Match')).toBe('"1"');
 expect(JSON.parse(String(patch.body))).toMatchObject({primaryLanguage: 'en', descriptionPolicy: 'managed_append', fields: ['description']});
 expect(request.mock.calls.some(([path]) => path.endsWith('/write-operations'))).toBe(false);
});

it('explains description and capability preview failures without publishing write authority', async () => {
 const user = userEvent.setup(); const draft = {...descriptionDraft(), state: 'staged' as const, fields: ['description' as const], primaryLanguage: 'uk', descriptionPolicy: 'replace' as const}; const detail = resultDetail(); detail.entry.draftId = draft.id;
 for (const [code, message] of [['DESCRIPTION_INVALID', /Description is not ready/], ['DESCRIPTION_TOO_LARGE', /Description exceeds the supported size/], ['BASELINE_REVIEW_REQUIRED', /Review current source and selected fields/], ['SOURCE_CHANGED', /source image changed.*Review current source and selected fields/], ['POLICY_CHANGED', /Writing policy changed/]]) {
  vi.stubGlobal('fetch', vi.fn(async (path: string) => path.includes('/write-previews') ? new Response(JSON.stringify({code}), {status: 409}) : new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : draft))));
  const view = render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} />);
  await user.click(await screen.findByRole('button', {name: 'Preview selected fields'}));
  expect(await screen.findByText(message, {selector: '[role="alert"]'})).toBeVisible();
  expect(screen.queryByRole('button', {name: 'Confirm selected fields write'})).not.toBeInTheDocument();
  expect(screen.getByLabelText('Description uk')).toHaveValue(draft.descriptions[0].text);
  view.unmount();
 }
});
