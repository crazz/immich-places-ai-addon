import {useEffect, useRef, useState} from 'react';

import {LaunchForm} from './LaunchForm';
import {previewSelection} from './selectionApi';
import {selectionIntent} from './selectionIntent';

import type {TJobProgress, TRerunChoice} from './jobTypes';
import type {TBatchIntention, TBatchScope} from './selectionIntent';
import type {TSelectionPreview} from './selectionTypes';
import type {ReactElement} from 'react';

type TBatchLaunchProps = {input: TBatchScope; onSubmittedAction: (job: TJobProgress) => void; rerun?: TRerunChoice};

function BatchLaunchSession({input, rerun, onSubmittedAction}: TBatchLaunchProps): ReactElement {
	const [preview, setPreview] = useState<TSelectionPreview | null>(null);
	const [error, setError] = useState('');
	const [isBusy, setBusy] = useState(false);
	const flight = useRef<AbortController | null>(null);
	useEffect(() => () => flight.current?.abort(), []);
	const runPreview = async (kind: TBatchIntention): Promise<void> => {
		flight.current?.abort();
		const controller = new AbortController();
		flight.current = controller;
		setBusy(true);
		setError('');
		setPreview(null);
		try {
			const result = await previewSelection(selectionIntent(kind, input), controller.signal);
			if (!controller.signal.aborted) {
				setPreview(result);
			}
		} catch (failure) {
			if (!controller.signal.aborted) {
				setError(failure instanceof Error ? failure.message : 'Could not preview this selection.');
			}
		} finally {
			if (!controller.signal.aborted) {
				setBusy(false);
			}
		}
	};
	return <section aria-label={'AI batch preview'} className={'space-y-3'}>
		<p>{'Preview the exact assets before authorizing image disclosure. Maximum 500 eligible assets.'}</p>
		{input.blockedReason && <p role={'alert'}>{input.blockedReason}</p>}
		<div className={'flex flex-wrap gap-2'}>
			<button type={'button'} disabled={isBusy || !!input.blockedReason} onClick={async () => runPreview('selected')}>{'Preview selected assets'}</button>
			{!rerun && <>
				<button type={'button'} disabled={isBusy || !!input.blockedReason} onClick={async () => runPreview('page')}>{'Preview current page'}</button>
				<button type={'button'} disabled={isBusy || !!input.blockedReason} onClick={async () => runPreview('matching')}>{'Preview all matching assets'}</button>
              </>}
		</div>
		{isBusy && <p role={'status'}>{'Preparing selection preview…'}</p>}
		{error && <p role={'alert'}>{error}</p>}
		{preview && <div className={'rounded border border-(--color-border) p-3'}>
			<p>{`Scope: ${preview.scope.view} · ${preview.mode === 'all-matching' ? 'All matching' : 'Explicit assets'}`}</p>
			<p>{`GPS: ${preview.scope.gpsFilter} · Visibility: ${preview.scope.hiddenFilter}`}</p>
			{preview.scope.albumID && <p>{`Album: ${preview.scope.albumID}`}</p>}
			{preview.scope.folderPath && <p>{`Folder: ${preview.scope.folderPath}`}</p>}
			{preview.scope.tagID && <p>{`Tag: ${preview.scope.tagID}`}</p>}
			{preview.scope.startDate && <p>{`From: ${preview.scope.startDate}`}</p>}
			{preview.scope.endDate && <p>{`Through: ${preview.scope.endDate}`}</p>}
			<p>{`Matched/requested: ${preview.matchedCount ?? preview.requestedCount} · Eligible: ${preview.eligibleCount} · Excluded: ${preview.excludedCount}`}</p>
			<p>{`Preview expires: ${new Date(preview.expiresAt).toLocaleString()}`}</p>
			{preview.exclusionCounts && <p>{Object.entries(preview.exclusionCounts).map(([reason, count]) => `${reason}: ${count}`).join(' · ')}</p>}
			{preview.exclusions && preview.exclusions.length > 0 && <p>{`Exclusions: ${[...new Set(preview.exclusions.map(item => item.reason))].join(', ')}`}</p>}
			{(!preview.snapshotID || preview.eligibleCount === 0) && <p>{'No eligible batch. Change the selection and preview again.'}</p>}
              </div>}
		{preview?.snapshotID && preview.eligibleCount > 0 && <LaunchForm key={preview.snapshotID} preview={preview} rerun={rerun} onSubmittedAction={onSubmittedAction} />}
        </section>;
}

export function BatchLaunch(props: TBatchLaunchProps): ReactElement {
	return <BatchLaunchSession key={JSON.stringify([props.input, props.rerun])} {...props} />;
}
