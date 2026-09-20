import {act, fireEvent, render, screen} from '@testing-library/react';
import {expect, it, vi} from 'vitest';

import {ResultDetail} from './ResultDetail';
import {fetchResult} from './resultsApi';
import {reviewDetail} from './testing/review';

vi.mock('./resultsApi', () => ({fetchResult: vi.fn()}));
vi.mock('./ReviewMapDynamic', () => ({ReviewMapDynamic: ({focus}: {focus: string | null}) => <output aria-label={'Map focus'}>{focus || 'all'}</output>}));

it('clears private inspection while changing results or owners and restores canonical focus and language', async () => {
 const first = reviewDetail(); const second = reviewDetail();
 second.entry.id = '22222222-2222-4222-8222-222222222222';
 second.entry.analysisId = '33333333-3333-4333-8333-333333333333';
 const deferred = Promise.withResolvers<typeof second>();
 vi.mocked(fetchResult).mockResolvedValueOnce(first).mockReturnValueOnce(deferred.promise).mockReturnValue(new Promise(() => {}));
 const view = render(<ResultDetail owner={'one'} reference={{analysisId: first.entry.analysisId!}} />);
 expect(await screen.findByLabelText('Map focus')).toHaveTextContent('candidate-1');
 fireEvent.click(screen.getByRole('button', {name: 'Show all alternatives'}));
 fireEvent.click(screen.getByRole('tab', {name: 'uk'}));
 expect(screen.getByLabelText('Map focus')).toHaveTextContent('all');
 expect(screen.getByRole('tab', {name: 'uk'})).toHaveAttribute('aria-selected', 'true');
 const signal = vi.mocked(fetchResult).mock.calls[0][1];
 view.rerender(<ResultDetail owner={'one'} reference={{analysisId: second.entry.analysisId!}} />);
 expect(signal?.aborted).toBe(true);
 expect(screen.queryByLabelText('Map focus')).not.toBeInTheDocument();
 await act(async () => deferred.resolve(second));
 expect(await screen.findByLabelText('Map focus')).toHaveTextContent('candidate-1');
 expect(screen.getByRole('tab', {name: 'en · Primary'})).toHaveAttribute('aria-selected', 'true');
 view.rerender(<ResultDetail owner={'two'} reference={{analysisId: second.entry.analysisId!}} />);
 expect(screen.queryByLabelText('Map focus')).not.toBeInTheDocument();
 expect(screen.queryByRole('tab')).not.toBeInTheDocument();
 view.unmount();
});
