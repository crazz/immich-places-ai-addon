import {fireEvent, render, screen, waitFor} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {fetchJobs} from './jobApi';
import {JobList} from './JobList';
import {aiProgress} from './testing/fixtures';

vi.mock('./jobApi', () => ({fetchJobs: vi.fn()}));
afterEach(() => vi.resetAllMocks());

it('opens saved jobs and follows only the returned pagination cursor without submitting work', async () => {
 const first = {...aiProgress(), items: null};
 const second = {...first, id: '55555555-5555-4555-8555-555555555555'};
 vi.mocked(fetchJobs).mockResolvedValueOnce({items: [first], nextCursor: 'opaque-cursor'}).mockResolvedValue({items: [second]});
 const open = vi.fn();
 render(<JobList owner={'owner'} onOpenAction={open} />);
 fireEvent.click(await screen.findByRole('button', {name: `Open job ${first.id}`}));
 expect(open).toHaveBeenCalledWith(first.id);
 fireEvent.click(screen.getByRole('button', {name: 'Older jobs'}));
 expect(await screen.findByRole('button', {name: `Open job ${second.id}`})).toBeVisible();
 await waitFor(() => expect(fetchJobs).toHaveBeenLastCalledWith('opaque-cursor', expect.any(AbortSignal)));
 expect(screen.queryByRole('button', {name: `Open job ${first.id}`})).not.toBeInTheDocument();
 expect(screen.queryByRole('button', {name: 'Older jobs'})).not.toBeInTheDocument();
});
