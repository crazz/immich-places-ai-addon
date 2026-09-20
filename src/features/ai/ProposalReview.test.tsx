import {fireEvent, render, screen} from '@testing-library/react';
import {expect, it, vi} from 'vitest';

import {ProposalReview} from './ProposalReview';
import {parseReview} from './reviewParser';
import {reviewDetail} from './testing/review';

vi.mock('./ReviewMapDynamic', () => ({ReviewMapDynamic: ({focus}: {focus: string | null}) => <output aria-label={'Map focus'}>{focus || 'all'}</output>}));
it('starts ambiguous inspection without a chosen candidate and changes only local focus', () => {
 const review = parseReview(reviewDetail()); if (!review) {throw new Error('Invalid fixture');}
 review.outcome = 'ambiguous'; review.initialCandidate = null; review.candidates.push({...review.candidates[0], id: 'candidate-2', name: 'Alternative viewpoint'});
 const before = JSON.stringify(review); render(<ProposalReview review={review} />);
 expect(screen.getByLabelText('Map focus')).toHaveTextContent('all');
 const inspect = screen.getByRole('button', {name: 'Inspect candidate candidate-2: Alternative viewpoint'});
 expect(inspect).toHaveAttribute('aria-pressed', 'false'); fireEvent.click(inspect);
 expect(inspect).toHaveAttribute('aria-pressed', 'true'); expect(screen.getByLabelText('Map focus')).toHaveTextContent('candidate-2');
 expect(screen.getByText('Description refers to candidate-1, not the currently inspected alternative.')).toBeVisible();
 expect(screen.getByText('Inspecting candidate-2. No approval has been recorded.')).toBeVisible();
 expect(screen.queryByRole('button', {name: /accept|save|apply/i})).not.toBeInTheDocument();
 expect(JSON.stringify(review)).toBe(before);
 fireEvent.click(screen.getByRole('button', {name: 'Show all alternatives'})); expect(screen.getByLabelText('Map focus')).toHaveTextContent('all');
});
