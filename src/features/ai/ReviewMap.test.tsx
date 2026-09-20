import {act, fireEvent, render, screen} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import * as mapBase from '@/features/map';

import {ReviewMap} from './ReviewMap';
import {parseReview} from './reviewParser';
import {reviewDetail} from './testing/review';

afterEach(() => vi.restoreAllMocks());
it('keeps tiles and inspection independent, fits only on focus, and removes private layers on unmount', () => {
 const review = parseReview(reviewDetail()); if (!review) {throw new Error('Invalid fixture');}
 review.candidates[0].camera!.radius = null; review.candidates[0].direction = null;
 const factory = vi.spyOn(mapBase, 'createBaseMap');
 const view = render(<ReviewMap review={review} focus={review.initialCandidate} />);
 expect(factory).toHaveBeenCalledTimes(1);
 const {map, tiles} = factory.mock.results[0].value;
 const fit = vi.spyOn(map, 'fitBounds'); const redraw = vi.spyOn(tiles, 'redraw'); const remove = vi.spyOn(map, 'remove');
 act(() => {tiles.fire('tileerror');});
 expect(screen.getByText('Map tiles unavailable. All coordinates and evidence remain available below.')).toBeVisible();
 fireEvent.click(screen.getByRole('button', {name: 'Retry map tiles'}));
 expect(redraw).toHaveBeenCalledTimes(1); expect(fit).not.toHaveBeenCalled();
 view.rerender(<ReviewMap review={review} focus={null} />);
 expect(fit).toHaveBeenCalledTimes(1); expect(factory).toHaveBeenCalledTimes(1);
 fireEvent.contextMenu(screen.getByRole('region', {name: 'Read-only proposal map'}));
 expect(screen.queryByText(/save location/i)).not.toBeInTheDocument();
 view.unmount(); expect(remove).toHaveBeenCalledTimes(1);
});
