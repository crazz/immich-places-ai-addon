import {isRecord} from '@/utils/typeGuards';

import {aiRequest} from './aiRequest';
import {isJobPage, isJobProgress} from './jobValidation';

import type {TJobAdmission, TJobPage, TJobProgress} from './jobTypes';

export async function submitJob(input: TJobAdmission, signal?: AbortSignal): Promise<TJobProgress> {
	return aiRequest('/ai/jobs', 'POST', input, isJobProgress, signal);
}

export async function fetchJob(id: string, signal?: AbortSignal): Promise<TJobProgress> {
	return aiRequest(`/ai/jobs/${encodeURIComponent(id)}`, 'GET', undefined, isJobProgress, signal);
}
export async function fetchJobs(cursor?: string, signal?: AbortSignal): Promise<TJobPage> {
	const query = cursor ? `?cursor=${encodeURIComponent(cursor)}` : '';
	return aiRequest(`/ai/jobs${query}`, 'GET', undefined, isJobPage, signal);
}
export async function cancelJob(id: string, signal?: AbortSignal): Promise<{canceled: true}> {
	return aiRequest(`/ai/jobs/${encodeURIComponent(id)}/cancel`, 'POST', {}, (value): value is {canceled: true} => isRecord(value) && value.canceled === true, signal);
}
