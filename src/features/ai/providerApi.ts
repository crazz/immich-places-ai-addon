import {backendFetch, parseJSON} from '@/shared/services/backendApi.fetch';
import {getBackendBaseURL} from '@/utils/backendUrls';
import {isRecord} from '@/utils/typeGuards';

import {isCapabilityReport} from './capabilityTypes';
import {raceAbort, withCompleteOperationDeadline} from './completeOperation';
import {isExecutionReadiness} from './executionReadiness';

import type {TCapabilityReport} from './capabilityTypes';
import type {TExecutionReadiness} from './executionReadiness';

export type {
	TCapabilityReport, TCapabilityObservation, TObservationStatus
} from './capabilityTypes';

export type TProviderInput = {
	name: string;
	baseURL: string;
	model: string;
	enabled: boolean;
	secret?: string;
};

export type TProviderProfile = Omit<TProviderInput, 'secret'> & {
	id: string;
	revision: number;
	hasSecret: boolean;
	capabilityReport?: TCapabilityReport;
	executionReadiness?: TExecutionReadiness;
};

export type TProviderList = {enabled: boolean; items: TProviderProfile[]};

export const CAPABILITY_TEST_TIMEOUT_MS = 130_000;

export class ProviderAPIError extends Error {
	constructor(message: string, readonly code: string) {
		super(message);
	}
}

function isProvider(value: unknown): value is TProviderProfile {
	if (!isRecord(value) || typeof value.id !== 'string' || typeof value.name !== 'string' ||
		typeof value.baseURL !== 'string' || typeof value.model !== 'string' ||
		typeof value.enabled !== 'boolean' || typeof value.hasSecret !== 'boolean' ||
		typeof value.revision !== 'number' || !Number.isSafeInteger(value.revision) || value.revision <= 0) {
		return false;
	}
	if (value.executionReadiness !== undefined && value.executionReadiness !== null && !isExecutionReadiness(value.executionReadiness)) {
		return false;
	}
	if (value.capabilityReport === undefined || value.capabilityReport === null) {
		return true;
	}
	return isCapabilityReport(value.capabilityReport);
}

function isProviderList(value: unknown): value is TProviderList {
	return isRecord(value) && typeof value.enabled === 'boolean' && Array.isArray(value.items) && value.items.every(isProvider);
}

async function checkResponse(response: Response): Promise<void> {
	if (response.ok) {
		return;
	}
	let payload: unknown;
	try {
		payload = await response.json();
	} catch {
		payload = null;
	}
	if (isRecord(payload) && typeof payload.message === 'string' && payload.message.length > 0 && typeof payload.code === 'string') {
		throw new ProviderAPIError(payload.message, payload.code);
	}
	throw new ProviderAPIError(`Provider settings request failed (${response.status})`, 'REQUEST_FAILED');
}

export async function fetchProviders(signal?: AbortSignal): Promise<TProviderList> {
	const response = await backendFetch(`${getBackendBaseURL()}/ai/providers`, {cache: 'no-store'}, {signal});
	await checkResponse(response);
	return parseJSON(response, isProviderList, 'Invalid provider list');
}

export async function saveProvider(input: TProviderInput, profile?: TProviderProfile): Promise<TProviderProfile> {
	const path = profile ? `/ai/providers/${encodeURIComponent(profile.id)}` : '/ai/providers';
	const response = await backendFetch(`${getBackendBaseURL()}${path}`, {
		method: profile ? 'PUT' : 'POST',
		headers: new Headers([['Content-Type', 'application/json']]),
		body: JSON.stringify({name: input.name, baseURL: input.baseURL, model: input.model, enabled: input.enabled, secret: input.secret, expectedRevision: profile?.revision})
	});
	await checkResponse(response);
	return parseJSON(response, isProvider, 'Invalid provider response');
}

export async function testProvider(profile: TProviderProfile, signal?: AbortSignal): Promise<TCapabilityReport> {
	return withCompleteOperationDeadline(CAPABILITY_TEST_TIMEOUT_MS, signal, async deadlineSignal => {
		const response = await backendFetch(`${getBackendBaseURL()}/ai/providers/${encodeURIComponent(profile.id)}/test`, {
			method: 'POST',
			headers: new Headers([['Content-Type', 'application/json']]),
			body: JSON.stringify({expectedRevision: profile.revision})
		}, {signal: deadlineSignal, timeoutMs: CAPABILITY_TEST_TIMEOUT_MS});
		await raceAbort(deadlineSignal, checkResponse(response));
		return raceAbort(deadlineSignal, parseJSON(response, isCapabilityReport, 'Invalid capability report'));
	});
}
