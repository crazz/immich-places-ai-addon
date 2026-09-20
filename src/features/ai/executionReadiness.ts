import {isRecord} from '@/utils/typeGuards';

export type TExecutionReadiness = {
	status: 'ready' | 'policy_required' | 'capability_required' | 'policy_violated' | 'unavailable';
	source?: string;
	policyID?: string;
	maxInputTokens?: number;
	maxOutputTokens?: number;
	maxRequestBytes?: number;
	maxImageBytes?: number;
	contextAllowed: boolean;
	costStatus: 'unknown' | 'estimated';
	currency?: string;
};

export function isExecutionReadiness(value: unknown): value is TExecutionReadiness {
	if (!isRecord(value) || !['ready', 'policy_required', 'capability_required', 'policy_violated', 'unavailable'].includes(String(value.status)) ||
		typeof value.contextAllowed !== 'boolean' || !['unknown', 'estimated'].includes(String(value.costStatus))) {
		return false;
	}
	if (value.status === 'ready' || value.status === 'capability_required' || value.status === 'policy_violated') {
		return (value.source === 'operator-attested' || value.source === 'application-defaults') && typeof value.policyID === 'string' && /^[a-f0-9]{64}$/.test(value.policyID) &&
			['maxInputTokens', 'maxOutputTokens', 'maxRequestBytes', 'maxImageBytes'].every(key => typeof value[key] === 'number' && Number.isSafeInteger(value[key]) && value[key] > 0) &&
			(value.costStatus === 'unknown' || (typeof value.currency === 'string' && /^[A-Z]{3}$/.test(value.currency)));
	}
	return true;
}
