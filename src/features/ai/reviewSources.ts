import {isRecord} from '@/utils/typeGuards';

import {safeReferenceURL} from './referenceLinks';
import {nullableText, reviewStrings} from './reviewGeometryParser';

import type {TResultDetail} from './resultDetailTypes';
import type {TReviewSource} from './reviewTypes';

export function readReviewSources(detail: TResultDetail): {sources: TReviewSource[]; omissions: string[]} | null {
 const context = detail.provenance.Context;
 if (detail.entry.mode === 'visual') {return context == null ? {sources: [], omissions: []} : null;}
 if (!isRecord(context) || context.Version !== 'context-v1' || !Array.isArray(context.Sources) || context.Sources.length > 9 || !reviewStrings(context.Omissions)) {return null;}
 const sources: TReviewSource[] = [];
 for (const source of context.Sources) {
  if (!isRecord(source) || typeof source.ID !== 'string' || !source.ID || typeof source.Kind !== 'string' || !['capture_time', 'selected_album', 'user_hint', 'nearby_locations'].includes(source.Kind) || typeof source.Lineage !== 'string') {return null;}
  sources.push({id: source.ID, kind: source.Kind, lineage: source.Lineage || 'unknown'});
 }
 if (detail.entry.mode === 'research' && detail.proposal?.schema_version === '2.0') {
  for (const source of detail.proposal?.sources ?? []) {
   if (!isRecord(source) || typeof source.id !== 'string' || typeof source.url !== 'string' || !nullableText(source.title) || typeof source.relevance !== 'string') {continue;}
   sources.push({id: source.id, kind: 'answer_reference', lineage: 'unknown', url: safeReferenceURL(source.url), title: source.title, relevance: source.relevance});
  }
 }
 return {sources, omissions: context.Omissions};
}
