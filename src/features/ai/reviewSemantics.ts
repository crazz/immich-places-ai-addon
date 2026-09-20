import type {TReview} from './reviewTypes';

export function validReview(review: TReview): boolean {
 const candidates = new Map(review.candidates.map(item => [item.id, item]));
 const observations = new Set(review.observations.map(item => item.id));
 const sources = new Set(review.sources.map(item => item.id));
 const sourceIdentities = review.mode === 'research' ? new Set(review.sources.map(item => `${item.kind === 'answer_reference' ? 'answer' : 'input'}:${item.id}`)) : sources;
 if (candidates.size !== review.candidates.length || observations.size !== review.observations.length || sourceIdentities.size !== review.sources.length) {return false;}
 if (review.outcome === 'located' ? !review.initialCandidate || !candidates.get(review.initialCandidate)?.camera : review.initialCandidate !== null) {return false;}
 if (review.outcome === 'ambiguous' && review.candidates.length < 2) {return false;}
 if (review.candidates.some(item => item.observations.some(id => !observations.has(id)) || (review.mode !== 'research' && item.sources.some(id => !sources.has(id))))) {return false;}
 if (review.languages.length !== review.descriptions.length || new Set(review.descriptions.map(item => item.language)).size !== review.languages.length || !review.languages.includes(review.primaryLanguage)) {return false;}
 return review.descriptions.every(item => review.languages.includes(item.language) &&
  (item.basis === 'scene_only' ? item.candidateId === null : item.candidateId !== null && candidates.has(item.candidateId)) &&
  (item.status === 'complete' ? item.text !== null && item.text.trim().length > 0 && item.reason === null : item.text === null && item.reason !== null && item.reason.trim().length > 0));
}
