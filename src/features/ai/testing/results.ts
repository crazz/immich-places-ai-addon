import type {TResultEntry} from '../resultTypes';

export function resultEntry(overrides: Partial<TResultEntry> = {}): TResultEntry {
	return {
		id: '11111111-1111-4111-8111-111111111111',
jobId: '22222222-2222-4222-8222-222222222222',
assetId: '33333333-3333-4333-8333-333333333333',
analysisId: '44444444-4444-4444-8444-444444444444',
		terminalAt: '2026-09-20T12:00:00Z',
mode: 'visual',
model: 'original-model',
label: 'An uncertain scene',
executionState: 'succeeded',
proposalOutcome: 'unknown',
reviewState: 'unreviewed',
writeState: 'not_requested',
captureDay: null,
albumId: null,
albumLabel: null,
sourceAvailable: false,
		...overrides
	};
}
