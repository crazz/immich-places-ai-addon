import {isRecord} from '@/utils/typeGuards';

import {nullableText, readCandidate} from './reviewGeometryParser';
import {validReview} from './reviewSemantics';
import {readReviewSources} from './reviewSources';

import type {TResultDetail} from './resultDetailTypes';
import type {TReview, TReviewDescription} from './reviewTypes';

function readDescription(value: unknown): TReviewDescription | null {
 if (!isRecord(value) || typeof value.language !== 'string' || (value.status !== 'complete' && value.status !== 'unavailable') || (value.basis !== 'scene_only' && value.basis !== 'candidate') || !nullableText(value.text) || !nullableText(value.candidate_id) || !nullableText(value.unavailable_reason)) {return null;}
 return {language: value.language, status: value.status, text: value.text, candidateId: value.candidate_id, reason: value.unavailable_reason, basis: value.basis};
}
export function parseReview(detail: TResultDetail): TReview | null {
 const proposal = detail.proposal;
 if (!proposal || !isRecord(proposal) || !nullableText(proposal.selected_candidate_id)) {return null;}
 const context = readReviewSources(detail);
 if (!context) {return null;}
 const candidates = proposal.candidates.map(value => readCandidate(value, detail.entry.mode === 'research')); const descriptions = proposal.descriptions.map(readDescription);
 if (candidates.some(value => !value) || descriptions.some(value => !value)) {return null;}
 const review: TReview = {outcome: proposal.outcome, initialCandidate: proposal.selected_candidate_id, candidates: candidates.filter(value => value !== null), descriptions: descriptions.filter(value => value !== null), observations: proposal.observations, sources: context.sources, omissions: context.omissions, languages: detail.provenance.Languages, primaryLanguage: detail.provenance.PrimaryLanguage, warnings: proposal.warnings, mode: detail.entry.mode};
 return validReview(review) ? review : null;
}
