import {expect, it} from 'vitest';

import {reviewGeometry} from './reviewGeometry';
import {parseReview} from './reviewParser';
import {canonicalProposal, reviewDetail} from './testing/review';

it('preserves ambiguous alternatives, subject-only unknowns and nullable precision without selecting or inventing a camera', () => {
 const detail = reviewDetail(); const proposal = canonicalProposal();
 proposal.candidates.push({...proposal.candidates[0], id: 'candidate-2'});
 Reflect.set(proposal, 'selected_candidate_id', null); proposal.outcome = 'ambiguous';
 Object.assign(detail.proposal!, proposal); detail.entry.proposalOutcome = 'ambiguous';
 const ambiguous = parseReview(detail);
 expect(ambiguous?.initialCandidate).toBeNull(); expect(ambiguous?.candidates).toHaveLength(2);
 expect(reviewGeometry(ambiguous!, null).markers.every(item => item.label.startsWith('Alternative'))).toBe(true);
 proposal.outcome = 'unknown'; proposal.candidates.pop();
 Reflect.set(proposal.candidates[0], 'camera_location', null); Reflect.set(proposal.candidates[0], 'camera_direction', null);
 Object.assign(detail.proposal!, proposal); detail.entry.proposalOutcome = 'unknown';
 const unknown = parseReview(detail);
 expect(unknown?.candidates[0].camera).toBeNull();
 expect(reviewGeometry(unknown!, null).markers.map(item => item.kind)).toEqual(['subject']);
});
