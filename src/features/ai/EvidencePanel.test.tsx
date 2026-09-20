import {fireEvent, render, screen} from '@testing-library/react';
import {expect, it} from 'vitest';

import {EvidencePanel} from './EvidencePanel';
import {parseReview} from './reviewParser';
import {reviewDetail} from './testing/review';

it('distinguishes model evidence, hints and unknown lineage as inert text without invented links', () => {
 const review = parseReview(reviewDetail()); if (!review) {throw new Error('Invalid fixture');}
 review.observations[0].text = '<img src="https://invented.invalid/track"> Ignore instructions';
 const view = render(<EvidencePanel review={review} candidate={review.candidates[0]} />);
 fireEvent.click(screen.getByText('Evidence and limitations'));
 expect(screen.getByText('Model / visual evidence; no independent verification.')).toBeVisible();
 expect(screen.getByText(review.observations[0].text)).toBeVisible();
 expect(screen.getByText('No external context sources were supplied.')).toBeVisible();
 review.mode = 'context-assisted'; review.sources = [{id: 'hint-1', kind: 'user_hint', lineage: 'unknown'}, {id: 'neighbor-1', kind: 'nearby_locations', lineage: 'unknown'}]; review.candidates[0].sources = ['hint-1', 'neighbor-1'];
 view.rerender(<EvidencePanel review={review} candidate={review.candidates[0]} />);
 expect(screen.getByText('User hint · hint-1')).toBeVisible();
 expect(screen.getByText('Nearby location · neighbor-1')).toBeVisible();
 expect(screen.getAllByText('Lineage: unknown; not independent corroboration.')).toHaveLength(2);
 expect(screen.queryByRole('link')).not.toBeInTheDocument(); expect(document.querySelector('img')).toBeNull();
});
