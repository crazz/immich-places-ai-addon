import {isRecord} from '@/utils/typeGuards';

import {isUUID} from './selectionTypes';

export type TResultEntry = {
	id: string; jobId: string; assetId: string; analysisId: string | null;
	terminalAt: string; mode: 'visual' | 'context-assisted' | 'research'; model: string; label: string | null;
	executionState: 'succeeded' | 'failed' | 'canceled'; proposalOutcome: 'located' | 'ambiguous' | 'unknown' | null;
	reviewState: 'unreviewed'; writeState: 'not_requested'; failure?: string;
	captureDay: string | null; albumId: string | null; albumLabel: string | null; sourceAvailable: boolean;
};
export type TResultPage = {items: TResultEntry[]; nextCursor?: string};
export type TResultFilters = {state?: string; outcome?: string; assetId?: string; jobId?: string; startDate?: string; endDate?: string; albumId?: string; undated?: string};
export type TResultQuery = TResultFilters & {limit?: number; cursor?: string};

export function isResultEntry(value: unknown): value is TResultEntry {
	if (!isRecord(value) || ![value.id, value.jobId, value.assetId].every(isUUID) ||
		(value.analysisId !== null && !isUUID(value.analysisId)) ||
		!['visual', 'context-assisted', 'research'].includes(String(value.mode)) || typeof value.model !== 'string' || value.model.length > 256 ||
		typeof value.terminalAt !== 'string' || !Number.isFinite(Date.parse(value.terminalAt)) ||
		value.reviewState !== 'unreviewed' || value.writeState !== 'not_requested' || typeof value.sourceAvailable !== 'boolean') {return false;}
	for (const key of ['label', 'captureDay', 'albumId', 'albumLabel']) {if (value[key] !== null && (typeof value[key] !== 'string' || value[key].length > 4096)) {return false;}}
	if (value.failure !== undefined && (typeof value.failure !== 'string' || value.failure.length > 64)) {return false;}
	if (value.executionState === 'succeeded') {return isUUID(value.analysisId) && ['located', 'ambiguous', 'unknown'].includes(String(value.proposalOutcome));}
	return (value.executionState === 'failed' || value.executionState === 'canceled') && value.analysisId === null && value.proposalOutcome === null;
}

export function isResultPage(value: unknown): value is TResultPage {
	if (!isRecord(value) || !Array.isArray(value.items) || value.items.length > 100 || !value.items.every(isResultEntry)) {return false;}
	return new Set(value.items.map(item => item.id)).size === value.items.length &&
		(value.nextCursor === undefined || (typeof value.nextCursor === 'string' && value.nextCursor.length <= 2048));
}
