import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import {ResultDetail} from './ResultDetail';
import {fetchResult} from './resultsApi';
import {resultDetail} from './testing/resultDetail';
import {stackIDs, stackPreview, stackReview,stagedPreviewDraft} from './testing/writePreview';

vi.mock('./resultsApi', () => ({fetchResult: vi.fn()}));
afterEach(() => {vi.resetAllMocks(); vi.unstubAllGlobals(); sessionStorage.clear();});

it('S18 P15 requires explicit resolution for each selected manual GPS overlap and preserves unrelated pending work', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const detail = resultDetail(); detail.entry.draftId = draft.id;
 const pending = [stackIDs[0], stackIDs[2]]; const preview = stackPreview(); vi.mocked(fetchResult).mockResolvedValue(detail);
 const request = vi.fn(async (path: string, init: RequestInit) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : path.endsWith('/stack-reviews') ? stackReview(JSON.parse(String(init.body)).targetIds) : path.endsWith('/write-previews') ? preview : draft))); vi.stubGlobal('fetch', request);
 const props = {owner: 'owner', reference: {analysisId: detail.entry.analysisId!}};
 const view = render(<ResultDetail {...props} manualPendingIDs={pending} />);
 await user.click(await screen.findByRole('button', {name: 'Review stack members'}));
 await user.click(await screen.findByRole('checkbox', {name: `Include photo ${stackIDs[0]}`}));
 expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeDisabled();
 expect(screen.getByRole('button', {name: 'Read selected target baselines'})).toBeDisabled();
 expect(screen.getByText(`Resolve pending manual GPS for ${stackIDs[0]} by explicitly saving or discarding that manual change.`)).toBeVisible();
 view.rerender(<ResultDetail {...props} manualPendingIDs={[stackIDs[2]]} />);
 await user.click(screen.getByRole('button', {name: 'Read selected target baselines'}));
 for (const id of [draft.assetId, stackIDs[0]]) {await user.click(await screen.findByRole('checkbox', {name: `Reviewed source and GPS for ${id}`}));}
 await user.click(screen.getByRole('button', {name: 'Preview exact GPS'}));
 expect(await screen.findByRole('button', {name: 'Confirm GPS write'})).toBeEnabled();
 view.rerender(<ResultDetail {...props} manualPendingIDs={pending} />);
 expect(screen.queryByRole('button', {name: 'Confirm GPS write'})).not.toBeInTheDocument();
 expect(screen.queryByText('Preview status: usable')).not.toBeInTheDocument();
 expect(pending).toEqual([stackIDs[0], stackIDs[2]]);
 expect(request.mock.calls.some(([path]) => path.includes('/location') || path.endsWith('/write-operations'))).toBe(false);
 const reviews = request.mock.calls.filter(([path]) => path.endsWith('/stack-reviews'));
 expect(reviews).toHaveLength(2); expect(JSON.parse(String(reviews[1][1].body)).targetIds).toEqual([draft.assetId, stackIDs[0]].sort());
});

it('checks every restored preview target against current manual pending GPS without rereading the stack', async () => {
 const draft = stagedPreviewDraft(); const detail = resultDetail(); detail.entry.draftId = draft.id; const preview = stackPreview();
 vi.mocked(fetchResult).mockResolvedValue(detail); sessionStorage.setItem(`ai-write-preview:owner:${draft.id}:${draft.revision}`, preview.plan.id);
 const request = vi.fn(async (path: string) => new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : path.includes('/write-previews/') ? preview : draft))); vi.stubGlobal('fetch', request);
 render(<ResultDetail owner={'owner'} reference={{analysisId: detail.entry.analysisId!}} manualPendingIDs={[stackIDs[0], stackIDs[2]]} />);
 expect(await screen.findByText(`Resolve pending manual GPS for ${stackIDs[0]} by explicitly saving or discarding that manual change.`)).toBeVisible();
 expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeDisabled();
 expect(screen.queryByRole('button', {name: 'Confirm GPS write'})).not.toBeInTheDocument();
 expect(request.mock.calls.some(([path]) => path.includes('/stack-reviews'))).toBe(false);
});

it('resets stack target selection on a new draft revision without retaining stale sibling conflicts', async () => {
 const user = userEvent.setup(); let draft = stagedPreviewDraft(); const detail = resultDetail(); detail.entry.draftId = draft.id; vi.mocked(fetchResult).mockResolvedValue(detail);
 const request = vi.fn(async (path: string, init: RequestInit) => {
  if (init.method === 'PATCH') {draft = {...draft, revision: 2};}
  return new Response(JSON.stringify(path.includes('?draftId=') ? {items: [], nextCursor: ''} : path.endsWith('/stack-reviews') ? stackReview() : draft));
 }); vi.stubGlobal('fetch', request);
 const props = {owner: 'owner', reference: {analysisId: detail.entry.analysisId!}};
 const view = render(<ResultDetail {...props} manualPendingIDs={[]} />);
 await user.click(await screen.findByRole('button', {name: 'Review stack members'})); await user.click(await screen.findByRole('checkbox', {name: `Include photo ${stackIDs[0]}`}));
 await user.click(screen.getByRole('button', {name: 'Save draft'})); await screen.findByText('Saved staged · Revision 2');
 view.rerender(<ResultDetail {...props} manualPendingIDs={[stackIDs[0]]} />);
 expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeEnabled();
 expect(screen.getByText('1 selected · Maximum 50 including the analyzed photo.')).toBeVisible();
});
