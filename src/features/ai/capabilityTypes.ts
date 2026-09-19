import {isRecord} from '@/utils/typeGuards';

export type TObservationStatus = 'supported' | 'unsupported' | 'unverified';

export type TCapabilityObservation = {
	status: TObservationStatus;
	reason?: string;
};

export type TCapabilityReport = {
	attemptID: string;
	profileID: string;
	revision: number;
	protocolVersion: string;
	policyFingerprint: string;
	lifecycle: string;
	startedAt: string;
	deadlineAt: string;
	completedAt?: string;
	requestedModel: string;
	reportedModel?: string;
	observations: {
		image: TCapabilityObservation;
		json: TCapabilityObservation;
		strict: TCapabilityObservation;
	};
	compatibility: string;
	applicable: boolean;
	inputMayBeConsumed?: boolean;
	usage?: {
		promptTokens?: number;
		completionTokens?: number;
		totalTokens?: number;
	};
};

const observationStatuses = new Set<TObservationStatus>(['supported', 'unsupported', 'unverified']);

function isObservation(value: unknown): value is TCapabilityObservation {
	return isRecord(value) && typeof value.status === 'string' && observationStatuses.has(value.status as TObservationStatus) &&
		(value.reason === undefined || typeof value.reason === 'string');
}

export function isCapabilityReport(value: unknown): value is TCapabilityReport {
	if (!isRecord(value) || typeof value.attemptID !== 'string' || typeof value.profileID !== 'string' ||
		typeof value.protocolVersion !== 'string' || typeof value.policyFingerprint !== 'string' ||
		typeof value.lifecycle !== 'string' || typeof value.startedAt !== 'string' ||
		typeof value.deadlineAt !== 'string' || typeof value.requestedModel !== 'string' ||
		typeof value.compatibility !== 'string' || typeof value.applicable !== 'boolean' ||
		typeof value.revision !== 'number' || !Number.isSafeInteger(value.revision) || value.revision < 1 ||
		!isRecord(value.observations) || !isObservation(value.observations.image) ||
		!isObservation(value.observations.json) || !isObservation(value.observations.strict)) {
		return false;
	}
	if (value.completedAt !== undefined && typeof value.completedAt !== 'string') {
		return false;
	}
	if (value.reportedModel !== undefined && typeof value.reportedModel !== 'string') {
		return false;
	}
	return true;
}
