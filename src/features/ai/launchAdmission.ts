import {isJobConfiguration} from './jobValidation';

import type {TContextClass, TJobAdmission, TJobConfiguration, TRerunChoice} from './jobTypes';
import type {TProviderProfile} from './providerApi';
import type {TSelectionPreview} from './selectionTypes';

export type TLaunchFields = {
	profileId: string; mode: 'visual' | 'context-assisted' | 'research'; format: 'strict' | 'json';
	languages: string; primaryLanguage: string; maxCalls: string; maxTokens: string; outputTokens: string; costCap: string;
	classes: TContextClass[]; hint: string;
};

export function launchDefaults(profile: TProviderProfile | undefined, count: number): Pick<TLaunchFields, 'profileId' | 'format' | 'maxCalls' | 'maxTokens' | 'outputTokens' | 'costCap'> {
	const ready = profile?.executionReadiness;
	const output = Math.min(4000, ready?.maxOutputTokens ?? 4000);
	return {
		profileId: profile?.id ?? '',
		format: profile?.capabilityReport?.observations.strict.status !== 'supported' && profile?.capabilityReport?.observations.json.status === 'supported' ? 'json' : 'strict',
		maxCalls: String(count),
maxTokens: String(((ready?.maxInputTokens ?? 100000) + output) * count),
outputTokens: String(output),
costCap: ''
	};
}

export function buildAdmission(fields: TLaunchFields, preview: Pick<TSelectionPreview, 'snapshotID' | 'eligibleCount' | 'expiresAt' | 'scope'>, profile: TProviderProfile, confirmed: boolean, key: string, rerun?: TRerunChoice): TJobAdmission {
	if (!confirmed) {
		throw new Error('Image disclosure consent is required.');
	}
	if (!preview.snapshotID || preview.eligibleCount < 1 || preview.eligibleCount > 500 || Date.parse(preview.expiresAt) <= Date.now()) {
		throw new Error('Create a fresh eligible selection preview.');
	}
	const ready = profile.executionReadiness;
	const observed = profile.capabilityReport;
	if (!profile.enabled) {
		throw new Error('Enable this provider in Settings → AI providers.');
	}
	if (ready?.status === 'policy_required') {
		throw new Error('An operator-configured restriction needs updating for this provider.');
	}
	if (ready?.status === 'policy_violated') {
		throw new Error('This provider exceeded its configured limits. Review the operator restriction before restarting.');
	}
	if (fields.profileId !== profile.id || !profile.enabled || ready?.status !== 'ready' || !ready.policyID || !ready.maxInputTokens || !ready.maxOutputTokens ||
		!observed?.applicable || observed.observations.image.status !== 'supported' || observed.observations[fields.format].status !== 'supported' ||
		(fields.mode !== 'visual' && !ready.contextAllowed)) {
		throw new Error('Test this provider in Settings → AI providers and choose a supported mode and output format.');
	}
	let languages: string[];
	let primaryLanguage: string;
	try {
		const tags = fields.languages.split(',').map(tag => tag.trim());
		languages = Intl.getCanonicalLocales(tags);
		primaryLanguage = Intl.getCanonicalLocales(fields.primaryLanguage.trim())[0];
		if (languages.length !== tags.length || !languages.includes(primaryLanguage)) {
			throw new Error('Invalid language set');
		}
	} catch {
		throw new Error('Use distinct valid language tags and a primary language from that set.');
	}
	const configuration: TJobConfiguration = {
		selectionToken: preview.snapshotID,
profileId: profile.id,
revision: profile.revision,
mode: fields.mode,
		format: fields.format,
allowJson: fields.format === 'json',
languages,
primaryLanguage,
policyId: ready.policyID,
		limits: {maxCalls: Number(fields.maxCalls), maxTokens: Number(fields.maxTokens), outputTokens: Number(fields.outputTokens)},
		...(rerun ? {rerun} : {})
	};
	if (fields.costCap.trim()) {
		const amount = fields.costCap.trim();
		if (!/^\d+(?:\.\d{1,6})?$/.test(amount)) {
			throw new Error('Use a nonnegative currency amount with at most six decimal places.');
		}
		const [whole, fraction = ''] = amount.split('.');
		configuration.limits.maxEstimatedMicros = Number(whole) * 1000000 + Number(fraction.padEnd(6, '0'));
		if (ready.costStatus !== 'estimated') {
			throw new Error('An estimated cost cap requires an attested tariff.');
		}
	}
	if (fields.mode !== 'visual') {
		configuration.context = {version: 'context-v1', classes: [...fields.classes]};
		if (fields.mode === 'research' && fields.hint !== '' && !configuration.context.classes.includes('user_hint')) {
			configuration.context.classes.push('user_hint');
		}
		if (configuration.context.classes.includes('user_hint')) {
			configuration.context.hint = fields.hint;
		}
		if (fields.classes.includes('selected_album')) {
			if (preview.scope.view !== 'album' || !preview.scope.albumID) {
				throw new Error('Album context requires a concrete selected album.');
			}
			configuration.context.albumId = preview.scope.albumID;
		}
	}
	if (!isJobConfiguration(configuration) || configuration.limits.maxCalls < 1 || configuration.limits.maxCalls > preview.eligibleCount * 3 || configuration.limits.outputTokens > ready.maxOutputTokens ||
		configuration.limits.maxTokens < ready.maxInputTokens + configuration.limits.outputTokens || !key) {
		throw new Error('Use finite calls, tokens and output limits within the displayed policy.');
	}
	return {configuration, consent: {version: 'image-consent-v1', image: true, configuration}, idempotencyKey: key};
}
