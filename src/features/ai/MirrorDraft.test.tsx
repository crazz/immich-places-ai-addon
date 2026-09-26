import {fireEvent, render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import {DraftReview} from './DraftReview';
import {parseReview} from './reviewParser';
import {mirroredDraft} from './testing/mirror';
import {resultDetail} from './testing/resultDetail';

afterEach(() => {vi.unstubAllGlobals(); sessionStorage.clear();});

it('R02 preserves unsaved mirror and text edits while an unresolved metadata write protects the revision', async () => {
 const user = userEvent.setup(); const draft = mirroredDraft(); const detail = resultDetail(); detail.entry.draftId = draft.id;
 const request = vi.fn(async (path: string, init: RequestInit) => init.method === 'PATCH' ? new Response(JSON.stringify({code: 'WRITE_IN_PROGRESS'}), {status: 409}) : new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : draft))); vi.stubGlobal('fetch', request);
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} />);
 fireEvent.change(await screen.findByLabelText('Description uk'), {target: {value: '  New local\ntext  '}}); await user.click(screen.getByLabelText('Mirror direction')); await user.click(screen.getByRole('button', {name: 'Save draft'}));
 expect(await screen.findByText('A standard or metadata operation can still act on this revision. Your edits are preserved; check its status before saving.')).toBeVisible();
 expect(screen.getByLabelText('Description uk')).toHaveValue('  New local\ntext  '); expect(screen.getByLabelText('Mirror direction')).not.toBeChecked(); expect(screen.getByText('Saved staged · Revision 1')).toBeVisible();
 expect(screen.getByRole('button', {name: 'Save draft'})).toBeDisabled(); expect(request.mock.calls.filter(([, init]) => init.method === 'PATCH')).toHaveLength(1); expect(request.mock.calls.some(([, init]) => init.method === 'POST')).toBe(false);
});

it('R01 accepts a revised mirror choice before dispatch without replaying queued approval', async () => {
 const {mirrorOperation} = await import('./testing/mirror'); const queued = mirrorOperation(); queued.status = 'queued'; queued.mirror.status = 'queued'; const user = userEvent.setup(); const draft = mirroredDraft(); const detail = resultDetail(); detail.entry.draftId = draft.id; const revised = {...draft, revision: 2, state: 'draft', mirror: {...draft.mirror, direction: false}};
 const request = vi.fn(async (path: string, init: RequestInit) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [{id: queued.id, status: queued.status, draftRevision: 1, approvedAt: queued.approvedAt}], nextCursor: ''} : path.includes('/write-operations/') ? queued : init.method === 'PATCH' ? revised : draft))); vi.stubGlobal('fetch', request);
 render(<DraftReview owner={'owner'} detail={detail} review={parseReview(detail)!} />);
 expect(await screen.findByText('Metadata step: queued')).toBeVisible(); await user.click(screen.getByLabelText('Mirror direction')); await user.click(screen.getByRole('button', {name: 'Save draft'}));
 expect(await screen.findByText('Saved draft · Revision 2')).toBeVisible(); expect(screen.getByLabelText('Mirror direction')).not.toBeChecked(); expect(screen.getByRole('button', {name: 'Preview exact GPS'})).toBeDisabled();
 const patches = request.mock.calls.filter(([, init]) => init.method === 'PATCH'); expect(patches).toHaveLength(1); expect(new Headers(patches[0][1].headers).get('If-Match')).toBe('"1"');
 expect(JSON.parse(String(patches[0][1].body)).mirror).toEqual(revised.mirror); expect(request.mock.calls.some(([, init]) => init.method === 'POST')).toBe(false);
});
