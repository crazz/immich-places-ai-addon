import {expect, it} from 'vitest';

import {parseReview} from './reviewParser';
import {canonicalProposal, reviewDetail} from './testing/review';

it('preserves separate camera and subject points including zero coordinates, radius and heading', () => {
 const proposal = canonicalProposal();
 proposal.candidates[0].camera_location.latitude = 0; proposal.candidates[0].camera_location.longitude = 0;
 proposal.candidates[0].camera_location.estimated_radius_m = 0; proposal.candidates[0].camera_direction.azimuth_deg = 0;
 const result = parseReview(reviewDetail(proposal));
 expect(result?.initialCandidate).toBe('candidate-1');
 expect(result?.candidates[0].camera).toMatchObject({latitude: 0, longitude: 0, radius: 0, radiusBasis: 'visual_estimate', granularity: 'area'});
 expect(result?.candidates[0].subject?.location).toEqual({latitude: 50.001, longitude: 14.001});
 expect(result?.candidates[0].direction).toEqual({degrees: 0, uncertainty: 35, method: 'visual_estimate'});
 expect(result?.candidates[0].observations).toEqual(['obs-1']);
 expect(result?.descriptions[0].candidateId).toBe('candidate-1');
});

it('fails closed for unresolved namespaces, invalid selected candidates and incomplete requested descriptions', () => {
 const baseline = canonicalProposal();
 const mutations: ((proposal: typeof baseline) => void)[] = [
  proposal => {proposal.candidates[0].evidence_refs = ['missing'];},
  proposal => {proposal.candidates[0].source_refs.push('https://invented.invalid');},
  proposal => {proposal.selected_candidate_id = 'missing';},
  proposal => {proposal.candidates.push({...proposal.candidates[0]});},
  proposal => {proposal.descriptions.pop();},
  proposal => {proposal.descriptions[1].language = 'en';},
  proposal => {proposal.descriptions[0].candidate_id = 'missing';},
  proposal => {proposal.descriptions[0].status = 'unavailable';},
  proposal => {proposal.outcome = 'ambiguous';}
 ];
 for (const mutate of mutations) {
  const proposal = canonicalProposal(); mutate(proposal);
  const detail = reviewDetail(); Object.assign(detail.proposal!, proposal);
  expect(parseReview(detail)).toBeNull();
 }
});

it('resolves contextual source IDs separately from model observations without upgrading unknown lineage', () => {
 const proposal = canonicalProposal(); proposal.candidates[0].source_refs.push('hint-1', 'neighbor-1');
 const detail = reviewDetail(proposal); detail.entry.mode = 'context-assisted'; detail.provenance.Mode = 'context-assisted';
 detail.provenance.Context = {Version: 'context-v1', Sources: [{ID: 'hint-1', Kind: 'user_hint', Lineage: ''}, {ID: 'neighbor-1', Kind: 'nearby_locations', Lineage: 'unknown'}], Omissions: ['Selected album was unavailable.']};
 const result = parseReview(detail);
 expect(result?.sources).toEqual([{id: 'hint-1', kind: 'user_hint', lineage: 'unknown'}, {id: 'neighbor-1', kind: 'nearby_locations', lineage: 'unknown'}]);
 expect(result?.observations[0].kind).toBe('visual');
 expect(result?.omissions).toEqual(['Selected album was unavailable.']);
 expect(result?.candidates[0].sources).toEqual(['hint-1', 'neighbor-1']);
});
