import {expect, it} from 'vitest';

import {reviewGeometry} from './reviewGeometry';
import {parseReview} from './reviewParser';
import {reviewDetail} from './testing/review';

it('frames antimeridian points locally and preserves canonical numeric coordinates at projection limits', () => {
 const review = parseReview(reviewDetail()); if (!review) {throw new Error('Invalid fixture');}
 const camera = review.candidates[0].camera!; const subject = review.candidates[0].subject!.location!;
 camera.latitude = 0; camera.longitude = 179.8; subject.latitude = 1; subject.longitude = -179.8;
 const geometry = reviewGeometry(review, review.initialCandidate);
 expect(geometry.markers.map(item => item.kind)).toEqual(['camera', 'subject']);
 expect(Math.abs(geometry.bounds[1].longitude - geometry.bounds[0].longitude)).toBeCloseTo(0.4);
 expect(subject.longitude).toBe(-179.8);
 camera.latitude = 90;
 const polar = reviewGeometry(review, review.initialCandidate);
 expect(polar.degraded).toBe(true); expect(polar.markers.map(item => item.kind)).toEqual(['subject']);
 expect(camera.latitude).toBe(90);
});
