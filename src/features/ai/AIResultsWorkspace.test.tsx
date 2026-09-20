import {fireEvent, render, screen, waitFor} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {AIResultsWorkspace} from './AIResultsWorkspace';
import {fetchResult, fetchResults} from './resultsApi';
import {resultDetail} from './testing/resultDetail';

vi.mock('./resultsApi', () => ({fetchResults: vi.fn(), fetchResult: vi.fn()}));
afterEach(() => {vi.resetAllMocks(); window.history.replaceState(null, '', '/');});
it('restores filters, exact page, scroll and row focus after detail, including reload navigation', async () => {
 const detail = resultDetail();
 vi.mocked(fetchResults).mockImplementation(async query => ({items: [detail.entry], nextCursor: query.cursor ? undefined : 'older'}));
 vi.mocked(fetchResult).mockResolvedValue(detail);
 const view = render(<AIResultsWorkspace owner={'owner'} />);
 fireEvent.click(screen.getByRole('button', {name: 'AI Results'}));
 fireEvent.click(screen.getByText('Filter results'));
 fireEvent.change(screen.getByLabelText('Outcome'), {target: {value: 'unknown'}});
 fireEvent.click(screen.getByRole('button', {name: 'Apply filters'}));
 fireEvent.click(await screen.findByRole('button', {name: 'Older results'}));
 await waitFor(() => expect(fetchResults).toHaveBeenLastCalledWith({outcome: 'unknown', cursor: 'older'}, expect.any(AbortSignal)));
 const list = screen.getByRole('region', {name: 'Results list'}); list.scrollTop = 120;
 const row = await screen.findByRole('button', {name: `Open result ${detail.entry.id}`}); row.focus(); fireEvent.click(row);
 expect(await screen.findByText('Proposal: unknown')).toBeVisible();
 expect(new URL(window.location.href).searchParams.get('aiResult')).toBe(detail.entry.analysisId);
 fireEvent.click(screen.getByRole('button', {name: 'Back to results'}));
 expect(screen.getByLabelText('Outcome')).toHaveValue('unknown');
 expect(list.scrollTop).toBe(120);
 await waitFor(() => expect(row).toHaveFocus());
 expect(fetchResults).toHaveBeenCalledTimes(3);
 fireEvent.click(row);
 await screen.findByText('Proposal: unknown');
 view.unmount(); render(<AIResultsWorkspace owner={'owner'} />);
 expect(await screen.findByText('Proposal: unknown')).toBeVisible();
 fireEvent.click(screen.getByRole('button', {name: 'Close dialog'}));
 await waitFor(() => expect(screen.getByRole('button', {name: 'AI Results'})).toHaveFocus());
 expect(new URL(window.location.href).searchParams.has('aiResult')).toBe(false);
});
