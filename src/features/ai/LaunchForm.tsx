import {useEffect, useRef, useState} from 'react';

import {submitJob} from './jobApi';
import {buildAdmission} from './launchAdmission';
import {LaunchFields} from './LaunchFields';
import {useProviderList} from './useProviderList';

import type {TJobAdmission, TJobProgress, TRerunChoice} from './jobTypes';
import type {TLaunchFields} from './launchAdmission';
import type {TProviderList} from './providerApi';
import type {TSelectionPreview} from './selectionTypes';
import type {ReactElement} from 'react';

type TLaunchProps = {preview: TSelectionPreview; onSubmittedAction: (job: TJobProgress) => void; rerun?: TRerunChoice};

function LaunchEditor({preview, providers, onSubmittedAction, rerun, reload}: TLaunchProps & {providers: TProviderList; reload: () => void}): ReactElement {
	const initial = providers.items.find(item => item.executionReadiness?.status === 'ready') ?? providers.items[0];
	const output = Math.min(4000, initial?.executionReadiness?.maxOutputTokens ?? 1);
	const [fields, setFields] = useState<TLaunchFields>({profileId: initial?.id ?? '', mode: 'visual', format: 'strict', languages: 'en', primaryLanguage: 'en', maxCalls: String(preview.eligibleCount), maxTokens: String(((initial?.executionReadiness?.maxInputTokens ?? 0) + output) * preview.eligibleCount), outputTokens: String(output), costCap: '', classes: [], hint: ''});
	const [isConsented, setConsented] = useState(false);
	const [pending, setPending] = useState<TJobAdmission | null>(null);
	const [isBusy, setBusy] = useState(false);
	const [isSubmitted, setSubmitted] = useState(false);
	const [failure, setFailure] = useState('');
	const [isExpired, setExpired] = useState(false);
	const flight = useRef<AbortController | null>(null);
	const locked = useRef(false);
	useEffect(() => {
		const timer = setTimeout(() => setExpired(true), Math.max(0, Date.parse(preview.expiresAt) - Date.now()));
		return () => { clearTimeout(timer); flight.current?.abort(); };
	}, [preview.expiresAt]);
	const profile = providers.items.find(item => item.id === fields.profileId);
	let validationError = '';
	try {
		if (!profile) { throw new Error('Configure a provider in AI providers.'); }
		buildAdmission(fields, preview, profile, true, 'validation', rerun);
	} catch (error) {
		validationError = error instanceof Error ? error.message : 'Check the launch configuration.';
	}
	const update = (change: Partial<TLaunchFields>): void => {
		setFields(current => ({...current, ...change})); setConsented(false); setPending(null); setFailure('');
	};
	const send = async (request: TJobAdmission): Promise<void> => {
		if (locked.current) { return; }
		locked.current = true;
		const controller = new AbortController(); flight.current = controller;
		setBusy(true); setPending(request); setFailure('');
		try {
			const job = await submitJob(request, controller.signal);
			if (!controller.signal.aborted) { setSubmitted(true); onSubmittedAction(job); }
		} catch (error) {
			if (!controller.signal.aborted) { setFailure(error instanceof Error ? error.message : 'Submission acknowledgement unavailable.'); }
		} finally {
			locked.current = false;
			if (!controller.signal.aborted) { setBusy(false); }
		}
	};
	const start = async (): Promise<void> => {
		if (!profile || pending || locked.current) { return; }
		try { await send(buildAdmission(fields, preview, profile, isConsented, crypto.randomUUID(), rerun)); }
		catch (error) { setFailure(error instanceof Error ? error.message : 'Check the launch configuration.'); }
	};
	const ready = profile?.executionReadiness;
	const isPartial = Number(fields.maxCalls) < preview.eligibleCount || Number(fields.maxTokens) < ((ready?.maxInputTokens ?? 0) + Number(fields.outputTokens)) * preview.eligibleCount;
	return <form aria-label={'AI analysis launch'} className={'space-y-3'} onSubmit={event => { event.preventDefault(); void start(); }}>
		<LaunchFields fields={fields} profiles={providers.items} preview={preview} disabled={isBusy || pending !== null} onChangeAction={update} />
		<p>{`Readiness: ${ready?.status ?? 'policy required'} · Input allowance per call: ${ready?.maxInputTokens ?? 'unknown'} · Cost: ${ready?.costStatus ?? 'unknown'} ${ready?.currency ?? ''}`}</p>
		{isPartial && <p>{'This allowance may complete only part of the batch.'}</p>}
		<p>{'Results are proposals for review. Reservations remain consumed after uncertain delivery; reported usage and charges may differ.'}</p>
		<label className={'block'}><input type={'checkbox'} checked={isConsented} disabled={isBusy || pending !== null} onChange={event => setConsented(event.target.checked)} />{`I consent to sending the selected images and chosen context to ${profile?.name ?? 'the selected provider'} for this exact run.`}</label>
		{(validationError || isExpired) && <p role={'alert'}>{isExpired ? 'Selection preview expired. Preview again.' : validationError}</p>}
		{failure && <p role={'alert'}>{`${failure} The run may already exist. Reconcile this submission before starting another run.`}</p>}
		{isBusy && <p role={'status'}>{'Submitting analysis…'}</p>}
		{isSubmitted && <p role={'status'}>{'Analysis accepted. Work continues after this window closes.'}</p>}
		{!pending && <button type={'submit'} disabled={!isConsented || !!validationError || isExpired || isBusy}>{'Start analysis'}</button>}
		{pending && !isSubmitted && <button type={'button'} disabled={isBusy} onClick={async () => send(pending)}>{'Reconcile submission'}</button>}
		{pending && <button type={'button'} disabled={isBusy} onClick={() => { setPending(null); setSubmitted(false); setConsented(false); setFailure(''); }}>{'Start a different run'}</button>}
		<button type={'button'} disabled={isBusy || pending !== null} onClick={reload}>{'Reload provider readiness'}</button>
        </form>;
}

export function LaunchForm(props: TLaunchProps): ReactElement {
	const {data, error, isLoading, reload} = useProviderList();
	if (isLoading) { return <p role={'status'}>{'Loading provider readiness…'}</p>; }
	if (error || !data) { return <div><p role={'alert'}>{error ?? 'Provider readiness unavailable.'}</p><button type={'button'} onClick={reload}>{'Reload provider readiness'}</button></div>; }
	if (!data.enabled) { return <p>{'AI execution is disabled. Existing jobs remain available for inspection and cancellation.'}</p>; }
	return <LaunchEditor key={JSON.stringify(data)} {...props} providers={data} reload={reload} />;
}
