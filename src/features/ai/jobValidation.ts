import {isRecord} from '@/utils/typeGuards';

import {isCount, isUUID} from './selectionTypes';

import type {TJobConfiguration, TJobPage, TJobProgress} from './jobTypes';

export const JOB_STATES = ['queued', 'running', 'retry_wait', 'blocked', 'succeeded', 'failed', 'canceled'] as const;

export function isJobConfiguration(value: unknown): value is TJobConfiguration {
	if (!isRecord(value) || !isUUID(value.selectionToken) || typeof value.profileId !== 'string' || !value.profileId || !isCount(value.revision) || value.revision < 1 ||
		!['visual', 'context-assisted'].includes(String(value.mode)) || !['strict', 'json'].includes(String(value.format)) || typeof value.allowJson !== 'boolean' ||
		!Array.isArray(value.languages) || value.languages.length < 1 || value.languages.length > 10 || !value.languages.every(tag => typeof tag === 'string' && tag.length > 0 && tag.length <= 64) ||
		typeof value.primaryLanguage !== 'string' || !value.languages.includes(value.primaryLanguage) || typeof value.policyId !== 'string' || !/^[a-f0-9]{64}$/.test(value.policyId) || !isRecord(value.limits)) {
		return false;
	}
	const limits = value.limits;
	if (!isCount(limits.maxCalls) || limits.maxCalls > 1500 || !isCount(limits.maxTokens) || limits.maxTokens < 1 || limits.maxTokens > 1e12 || !isCount(limits.outputTokens) || limits.outputTokens < 1 || limits.outputTokens > 1e9 ||
		(limits.maxEstimatedMicros !== undefined && !isCount(limits.maxEstimatedMicros))) {
		return false;
	}
	if (value.rerun !== undefined && (!isRecord(value.rerun) || !isUUID(value.rerun.parentJobId) || !['retry-failed', 'reanalysis'].includes(String(value.rerun.kind)))) {
		return false;
	}
	if (value.mode === 'visual') {
		return value.context === undefined;
	}
	const context = value.context;
	return isRecord(context) && context.version === 'context-v1' && Array.isArray(context.classes) && context.classes.length <= 4 && new Set(context.classes).size === context.classes.length &&
		context.classes.every(item => ['capture_time', 'selected_album', 'user_hint', 'nearby_locations'].includes(String(item))) &&
		(context.hint === undefined || (typeof context.hint === 'string' && new TextEncoder().encode(context.hint).length <= 2000 && context.classes.includes('user_hint'))) &&
		(context.albumId === undefined || (typeof context.albumId === 'string' && context.classes.includes('selected_album')));
}

function validJob(value: unknown, summary: boolean): boolean {
	if (!isRecord(value) || !isUUID(value.id) || !isJobConfiguration(value.configuration) || typeof value.createdAt !== 'number' || !Number.isFinite(value.createdAt) || value.createdAt <= 0 ||
		typeof value.canceled !== 'boolean' || typeof value.blocked !== 'boolean' || !isRecord(value.counts) || !isCount(value.counts.total) || value.counts.total > 500 ||
		!Object.entries(value.counts).every(([key, count]) => (key === 'total' || JOB_STATES.some(state => state === key)) && isCount(count)) || !isRecord(value.usage)) {
		return false;
	}
	const usage = value.usage;
	const counts = value.counts;
	if (JOB_STATES.reduce((sum, state) => sum + Number(counts[state] ?? 0), 0) !== counts.total) {
		return false;
	}
	if (!isCount(usage.calls) || !isCount(usage.reservedTokens) || !['unknown', 'partial', 'complete'].includes(String(usage.reportedStatus)) || !['unknown', 'estimated'].includes(String(usage.costStatus)) ||
		!['inputReported', 'outputReported', 'totalReported', 'estimatedMicros'].every(key => usage[key] === null || isCount(usage[key]))) {
		return false;
	}
	if (summary) {
		return value.items === null;
	}
	const items = value.items;
	if (!Array.isArray(items) || items.length !== counts.total || !items.every(item => isRecord(item) && isUUID(item.id) && isUUID(item.assetId) && JOB_STATES.some(state => state === item.state) &&
		isCount(item.attempts) && item.attempts <= 3 && isCount(item.calls) && item.calls <= 3 && (item.resultId === null || isUUID(item.resultId)) &&
		(item.failure === undefined || typeof item.failure === 'string') && (item.outcome === undefined || ['located', 'ambiguous', 'unknown'].includes(String(item.outcome))))) {
		return false;
	}
	return new Set(items.map(item => item.id)).size === items.length && new Set(items.map(item => item.assetId)).size === items.length &&
		JOB_STATES.every(state => items.filter(item => item.state === state).length === (counts[state] ?? 0));
}
export function isJobProgress(value: unknown): value is TJobProgress {
	return validJob(value, false);
}
export function isJobPage(value: unknown): value is TJobPage {
	return isRecord(value) && Array.isArray(value.items) && value.items.length <= 100 && value.items.every(item => validJob(item, true)) && (value.nextCursor === undefined || value.nextCursor === null || (typeof value.nextCursor === 'string' && value.nextCursor.length <= 2048));
}
