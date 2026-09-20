import {render, screen} from '@testing-library/react';
import {expect, it} from 'vitest';

import {CandidateFacts} from './CandidateFacts';
import {parseReview} from './reviewParser';
import {reviewDetail} from './testing/review';

it('labels original camera/subject, uncalibrated zero radius and true-north zero heading without inventing null geometry', () => {
 const review = parseReview(reviewDetail()); if (!review) {throw new Error('Invalid fixture');}
 const candidate = review.candidates[0]; candidate.camera = {latitude: 0, longitude: 0, granularity: 'area', radius: 0, radiusBasis: 'visual_estimate'}; candidate.direction = {degrees: 0, uncertainty: 120, method: 'visual_estimate'};
 const view = render(<CandidateFacts candidate={candidate} />);
 expect(screen.getByText('Camera: 0, 0')).toBeVisible();
 expect(screen.getByText('Subject: 50.001, 14.001')).toBeVisible();
 expect(screen.getByText(/Estimated radius: 0 m · visual estimate · uncalibrated/)).toBeVisible();
 expect(screen.getByText('Heading: 0° clockwise from true north')).toBeVisible();
 expect(screen.getByText('Uncertainty: 120° · Method: visual estimate')).toBeVisible();
 view.rerender(<CandidateFacts candidate={{...candidate, camera: {...candidate.camera, radius: null, radiusBasis: 'unknown'}, direction: null}} />);
 expect(screen.getByText('Precision: unknown')).toBeVisible(); expect(screen.getByText('Heading: Unknown')).toBeVisible();
 view.rerender(<CandidateFacts candidate={{...candidate, camera: null, direction: null}} />);
 expect(screen.getByText('Camera: Unknown')).toBeVisible(); expect(screen.getByText('Subject: 50.001, 14.001')).toBeVisible();
});
