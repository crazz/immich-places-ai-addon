import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {expect, it} from 'vitest';

import {DescriptionTabs} from './DescriptionTabs';
import {parseReview} from './reviewParser';
import {reviewDetail} from './testing/review';

it('switches requested non-English tabs by keyboard and preserves unavailable and original candidate basis', async () => {
 const review = parseReview(reviewDetail()); if (!review) {throw new Error('Invalid fixture');}
 review.languages = ['fr', 'pt']; review.primaryLanguage = 'fr';
 review.descriptions = [{language: 'fr', status: 'complete', text: 'Une promenade au bord de la rivière.', basis: 'candidate', candidateId: 'candidate-1', reason: null}, {language: 'pt', status: 'unavailable', text: null, basis: 'scene_only', candidateId: null, reason: 'Requested translation was unavailable.'}];
 const user = userEvent.setup(); const view = render(<DescriptionTabs review={review} focus={'candidate-2'} />);
 expect(screen.getAllByRole('tab').map(item => item.textContent)).toEqual(['fr · Primary', 'pt']);
 expect(screen.getByText('Une promenade au bord de la rivière.')).toHaveAttribute('lang', 'fr');
 expect(screen.getByText('Description refers to candidate-1, not the currently inspected alternative.')).toBeVisible();
 screen.getByRole('tab', {name: 'fr · Primary'}).focus(); await user.keyboard('{ArrowRight}');
 expect(screen.getByRole('tab', {name: 'pt'})).toHaveFocus();
 expect(screen.getByText('Status: unavailable')).toBeVisible();
 expect(screen.getByText('Requested translation was unavailable.')).toBeVisible();
 expect(screen.getByText('Basis: visible scene only')).toBeVisible();
 document.documentElement.lang = 'en'; view.rerender(<DescriptionTabs review={review} focus={null} />);
 expect(screen.getAllByRole('tab')).toHaveLength(2);
 await user.keyboard('{Home}'); expect(screen.getByRole('tab', {name: 'fr · Primary'})).toHaveFocus();
});
