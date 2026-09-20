import {aiRequest} from './aiRequest';
import {isResultDetail} from './resultDetailTypes';
import {isResultPage} from './resultTypes';
import {isUUID} from './selectionTypes';

import type {TResultDetail, TResultReference} from './resultDetailTypes';
import type {TResultPage, TResultQuery} from './resultTypes';

export async function fetchResults(query: TResultQuery, signal?: AbortSignal): Promise<TResultPage> {
	const params = new URLSearchParams();
	for (const [key, value] of Object.entries(query)) {if (value !== undefined && value !== '') {params.set(key, String(value));}}
	return aiRequest(`/ai/results${params.size ? `?${params}` : ''}`, 'GET', undefined, isResultPage, signal);
}

export async function fetchResult(reference: TResultReference, signal?: AbortSignal): Promise<TResultDetail> {
	if ('analysisId' in reference ? !isUUID(reference.analysisId) : !isUUID(reference.jobId) || !isUUID(reference.itemId)) {throw new Error('Result unavailable');}
	const path = 'analysisId' in reference ? `/ai/results/${reference.analysisId}` : `/ai/jobs/${reference.jobId}/items/${reference.itemId}/result`;
	const detail = await aiRequest(path, 'GET', undefined, isResultDetail, signal);
	if ('analysisId' in reference ? detail.entry.analysisId !== reference.analysisId : detail.entry.jobId !== reference.jobId || detail.entry.id !== reference.itemId) {throw new Error('Result unavailable');}
	return detail;
}
