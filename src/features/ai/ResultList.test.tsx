import {fireEvent, render, screen} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {ResultList} from './ResultList';
import {fetchResults} from './resultsApi';
import {resultEntry} from './testing/results';

vi.mock('./resultsApi', () => ({fetchResults: vi.fn()}));
afterEach(() => vi.resetAllMocks());
it('shows truthful separate states, inert labels and exact cursor navigation without hiding prior runs', async () => {
 const entry = resultEntry({label: '<script>Untrusted</script>'});
 const failed = resultEntry({id: '55555555-5555-4555-8555-555555555555', executionState: 'failed', proposalOutcome: null, analysisId: null, label: null});
 vi.mocked(fetchResults).mockResolvedValue({items: [entry, failed], nextCursor: 'opaque'});
 const open = vi.fn(); const page = vi.fn();
 render(<ResultList owner={'owner'} query={{state: 'succeeded'}} onOpenAction={open} onPageAction={page} />);
 expect(await screen.findByText(entry.label!)).toBeVisible();
 expect(screen.getByText('Execution: succeeded · Proposal: unknown')).toBeVisible();
 expect(screen.getByText('Execution: failed · Proposal: none')).toBeVisible();
 expect(screen.getAllByText('Review: unreviewed · Write: not requested')).toHaveLength(2);
 expect(screen.getAllByText('Capture date unknown')).toHaveLength(2);
 expect(document.querySelector('script')).toBeNull();
 fireEvent.click(screen.getByRole('button', {name: `Open result ${entry.id}`}));
 expect(open).toHaveBeenCalledWith(entry);
 fireEvent.click(screen.getByRole('button', {name: 'Older results'}));
 expect(page).toHaveBeenCalledWith('opaque');
 expect(fetchResults).toHaveBeenCalledWith({state: 'succeeded'}, expect.any(AbortSignal));
});

it('offers explicit bounded retry after a failed history request and describes an empty filtered result', async () => {
 vi.mocked(fetchResults).mockRejectedValueOnce(new Error('offline')).mockResolvedValue({items: []});
 render(<ResultList owner={'owner'} query={{state: 'failed'}} onOpenAction={() => {}} onPageAction={() => {}} />);
 expect(await screen.findByRole('alert')).toHaveTextContent('Could not refresh');
 expect(fetchResults).toHaveBeenCalledTimes(1);
 fireEvent.click(screen.getByRole('button', {name: 'Refresh results'}));
 expect(await screen.findByText('No results match these filters.')).toBeVisible();
 expect(fetchResults).toHaveBeenCalledTimes(2);
});

it('shows the saved review state and directs draft owners to GPS history without claiming no write was requested', async () => {
 const entry = resultEntry({draftId: '66666666-6666-4666-8666-666666666666', draftRevision: 4, reviewState: 'staged'});
 vi.mocked(fetchResults).mockResolvedValue({items: [entry]});
 render(<ResultList owner={'owner'} query={{}} onOpenAction={() => {}} onPageAction={() => {}} />);
 expect(await screen.findByText('Review: staged · GPS status in result')).toBeVisible();
 expect(screen.queryByText(/Write: not requested/)).not.toBeInTheDocument();
});
