import type {TCapabilityReport, TProviderProfile} from './providerApi';
import type {ReactElement} from 'react';

const explainedCompatibility = new Set(['unsupported', 'incomplete', 'failed']);

function observationLabel(name: string, status: string): string {
	return `${name}: ${status}`;
}

function isStaleEvidence(report: TCapabilityReport, profile: TProviderProfile): boolean {
	if (report.revision !== profile.revision) {
		return true;
	}
	return !report.applicable && report.lifecycle === 'completed' && report.observations.image.status === 'supported';
}

function needsAction(report: TCapabilityReport, profile: TProviderProfile): 'retry' | 'reload' | null {
	if (report.lifecycle === 'interrupted') {
		return 'reload';
	}
	if (report.lifecycle === 'failed' || report.compatibility === 'unsupported' || report.compatibility === 'incomplete' ||
		report.compatibility === 'failed' || !report.applicable || report.revision !== profile.revision) {
		return 'retry';
	}
	return null;
}

export function ProviderCapabilityResult({
	profile,
	liveReport,
	isPending,
	error,
	errorCode,
	onRetryAction,
	onReloadAction
}: {
	profile: TProviderProfile;
	liveReport: TCapabilityReport | null;
	isPending: boolean;
	error: string | null;
	errorCode: string | null;
	onRetryAction: () => void;
	onReloadAction: () => void;
}): ReactElement | null {
	const report = liveReport ?? profile.capabilityReport ?? null;
	if (!isPending && !error && !report) {
		return null;
	}
	const isStale = report !== null && isStaleEvidence(report, profile);
	const action = report && !isPending && !error ? needsAction(report, profile) : null;
	return (
		<div className={'mt-2 space-y-1 text-xs'}>
			{isPending && <p role={'status'}>{'Testing provider…'}</p>}
			{error && (
				<>
					<p role={'alert'}>{error}</p>
					{errorCode === 'PROVIDER_BUSY' ? (
						<button type={'button'} className={'underline'} onClick={onReloadAction}>{'Reload report'}</button>
					) : (
						<button type={'button'} className={'underline'} onClick={onRetryAction}>{'Retry test'}</button>
					)}
				</>
			)}
			{report && !isPending && (
				<div role={'status'} className={'space-y-1'}>
					{isStale && <p>{'Stale: result is not current for this profile revision or policy.'}</p>}
					{report.lifecycle === 'interrupted' && <p>{'Interrupted: reload the report before starting another test.'}</p>}
					{report.lifecycle === 'canceled' && <p>{'Canceled: the test stopped before completing its observations.'}</p>}
					{report.lifecycle === 'failed' && <p>{`Failed: ${report.compatibility}`}</p>}
					{report.compatibility === 'unsupported' && <p>{'Unsupported for this protocol.'}</p>}
					{report.compatibility === 'incomplete' && <p>{'Unverified: incomplete observations.'}</p>}
					{report.lifecycle === 'completed' && !explainedCompatibility.has(report.compatibility) &&
						<p>{`Compatibility: ${report.compatibility}`}</p>}
					<p>{observationLabel('Image', report.observations.image.status)}</p>
					<p>{observationLabel('JSON', report.observations.json.status)}</p>
					<p>{observationLabel('Strict', report.observations.strict.status)}</p>
					<p>{`Tested revision ${report.revision}${report.completedAt ? ` · ${report.completedAt}` : ''}`}</p>
					<p className={'text-(--color-text-secondary)'}>
						{'Observed synthetic compatibility only — not geolocation quality or private-photo authorization.'}
					</p>
					{action === 'reload' && (
						<button type={'button'} className={'underline'} onClick={onReloadAction}>{'Reload report'}</button>
					)}
					{action === 'retry' && (
						<button type={'button'} className={'underline'} onClick={onRetryAction}>{'Retry test'}</button>
					)}
				</div>
			)}
		</div>
	);
}
