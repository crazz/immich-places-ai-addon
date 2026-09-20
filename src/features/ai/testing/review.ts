import synthetic from '../../../../docs/ai-locate/examples/ai-analysis-result.synthetic.json';
import {isResultDetail} from '../resultDetailTypes';
import {resultDetail} from './resultDetail';

import type {TResultDetail} from '../resultDetailTypes';

export function reviewDetail(proposal: unknown = structuredClone(synthetic)): TResultDetail {
 const base = resultDetail();
 const value = {...base, entry: {...base.entry, proposalOutcome: 'located'}, proposal, provenance: {...base.provenance, Languages: ['en', 'uk']}};
 if (!isResultDetail(value)) {throw new Error('Invalid review fixture');}
 return value;
}
type TCanonicalFixture = Omit<typeof synthetic, 'candidates'> & {candidates: (Omit<typeof synthetic.candidates[number], 'source_refs'> & Record<'source_refs', string[]>)[]};
export function canonicalProposal(): TCanonicalFixture {return structuredClone(synthetic);}
