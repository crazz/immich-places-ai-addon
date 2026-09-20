export type TContextClass = 'capture_time' | 'selected_album' | 'user_hint' | 'nearby_locations';
export type TContextChoices = {version: string; classes: TContextClass[]; hint?: string; albumId?: string};
export type TRerunChoice = {parentJobId: string; kind: 'retry-failed' | 'reanalysis'};
export type TJobConfiguration = {
	selectionToken: string;
	profileId: string;
	revision: number;
	mode: 'visual' | 'context-assisted' | 'research';
	format: 'strict' | 'json';
	allowJson: boolean;
	languages: string[];
	primaryLanguage: string;
	policyId: string;
	limits: {maxCalls: number; maxTokens: number; outputTokens: number; maxEstimatedMicros?: number};
	context?: TContextChoices;
	rerun?: TRerunChoice;
};
export type TJobAdmission = {configuration: TJobConfiguration; consent: {version: string; image: boolean; configuration: TJobConfiguration}; idempotencyKey: string};
export type TJobItem = {
	id: string; assetId: string;
	state: 'queued' | 'running' | 'retry_wait' | 'blocked' | 'succeeded' | 'failed' | 'canceled';
	failure?: string; attempts: number; calls: number; resultId: string | null;
	outcome?: 'located' | 'ambiguous' | 'unknown';
};
export type TJobUsage = {
	calls: number; reservedTokens: number;
	reportedStatus: 'unknown' | 'partial' | 'complete'; costStatus: 'unknown' | 'estimated';
	inputReported: number | null; outputReported: number | null; totalReported: number | null;
	estimatedMicros: number | null; currency?: string;
};
export type TJobProgress = {
	id: string; configuration: TJobConfiguration; createdAt: number; canceled: boolean; blocked: boolean;
	counts: Record<string, number>; items: TJobItem[]; usage: TJobUsage;
};
export type TJobSummary = Omit<TJobProgress, 'items'> & {items: null};
export type TJobPage = {items: TJobSummary[]; nextCursor?: string | null};
