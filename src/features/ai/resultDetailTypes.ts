import {isRecord} from '@/utils/typeGuards';

import {isResultEntry} from './resultTypes';

import type {TResultEntry} from './resultTypes';

export type TResultProvenance = {
	Mode: string; Installation: string; Profile: string; Model: string; Revision: number;
	Languages: string[]; PrimaryLanguage: string; SelectionDigest: string;
	SourceDigest: string; ImageDigest: string; PromptVersion: string; SchemaVersion: string; ValidationVersion: string;
	Context?: unknown;
};
type TProposalContent = Record<'selected_candidate_id', string | null> & {
	outcome: 'located' | 'ambiguous' | 'unknown';
	observations: {id: string; kind: string; text: string}[];
	warnings: string[]; candidates: unknown[]; descriptions: unknown[];
};
export type TStoredProposal = TProposalContent & (Record<'schema_version', '1.0'> | (Record<'schema_version', '2.0'> & {sources?: unknown[]}));
export type TResultDetail = {entry: TResultEntry; proposal: TStoredProposal | null; provenance: TResultProvenance};
export type TResultReference = {analysisId: string} | {jobId: string; itemId: string};

export function isResultDetail(value: unknown): value is TResultDetail {
	if (!isRecord(value) || !isResultEntry(value.entry) || !isRecord(value.provenance)) {return false;}
	const meta = value.provenance;
	if (!['Mode', 'Installation', 'Profile', 'Model', 'PrimaryLanguage', 'SelectionDigest', 'SourceDigest', 'ImageDigest', 'PromptVersion', 'SchemaVersion', 'ValidationVersion'].every(key => typeof meta[key] === 'string') ||
		meta.Model !== value.entry.model || (meta.Mode !== value.entry.mode && !(meta.Mode === '' && value.entry.mode === 'visual')) ||
		!Number.isSafeInteger(meta.Revision) || Number(meta.Revision) < 1 || !Array.isArray(meta.Languages) || meta.Languages.length < 1 || !meta.Languages.every(tag => typeof tag === 'string') || typeof meta.PrimaryLanguage !== 'string' || !meta.Languages.includes(meta.PrimaryLanguage)) {return false;}
	if (value.entry.executionState !== 'succeeded') {return value.proposal === null;}
	const proposal = value.proposal;
	const schema = value.entry.mode === 'research' ? '2.0' : '1.0';
	return isRecord(proposal) && proposal.schema_version === schema && (proposal.sources === undefined || (schema === '2.0' && Array.isArray(proposal.sources) && proposal.sources.length <= 16)) && (proposal.selected_candidate_id === null || typeof proposal.selected_candidate_id === 'string') && proposal.outcome === value.entry.proposalOutcome &&
		Array.isArray(proposal.observations) && proposal.observations.every(item => isRecord(item) && ['id', 'kind', 'text'].every(key => typeof item[key] === 'string')) &&
		Array.isArray(proposal.warnings) && proposal.warnings.every(item => typeof item === 'string') &&
		Array.isArray(proposal.candidates) && Array.isArray(proposal.descriptions);
}
