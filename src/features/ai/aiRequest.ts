import {backendFetch, parseJSON} from '@/shared/services/backendApi.fetch';
import {getBackendBaseURL} from '@/utils/backendUrls';
import {isRecord} from '@/utils/typeGuards';

import {raceAbort, withCompleteOperationDeadline} from './completeOperation';
import {isCount} from './selectionTypes';

export class AIRequestError extends Error {
	constructor(readonly code: string, message: string) {
		super(message);
	}
}

export async function aiRequest<T>(path: string, method: 'GET' | 'POST', input: unknown, validate: (value: unknown) => value is T, signal?: AbortSignal): Promise<T> {
	return withCompleteOperationDeadline(15_000, signal, async operationSignal => raceAbort(operationSignal, (async () => {
		const response = await backendFetch(`${getBackendBaseURL()}${path}`, {
			method,
			cache: 'no-store',
			headers: new Headers([['Content-Type', 'application/json']]),
			body: method === 'POST' ? JSON.stringify(input) : undefined
		}, {signal: operationSignal, timeoutMs: 15_000});
		if (!response.ok) {
			let failure: unknown;
			try { failure = await response.json(); } catch { failure = null; }
			const code = isRecord(failure) && typeof failure.code === 'string' ? failure.code : 'REQUEST_FAILED';
			if (code === 'SELECTION_LIMIT_EXCEEDED' && isRecord(failure) && ['matchedCount', 'eligibleCount', 'excludedCount'].every(key => isCount(failure[key]))) {
				throw new AIRequestError(code, `No batch was created. Maximum 500 eligible assets. Matched: ${failure.matchedCount} · Eligible: ${failure.eligibleCount} · Excluded: ${failure.excludedCount}. Narrow the selection and preview again.`);
			}
			throw new AIRequestError(code, code === 'AI_DISABLED' ? 'AI execution is disabled.' : 'The request could not be completed. Refresh the selection or reconcile the same submission.');
		}
		return parseJSON(response, validate, 'Invalid AI response');
	})()));
}
