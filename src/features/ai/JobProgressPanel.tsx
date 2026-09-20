import {useEffect, useRef, useState} from 'react';

import {cancelJob} from './jobApi';
import {jobIsActive, useJobProgress} from './useJobProgress';

import type {TJobItem, TRerunChoice} from './jobTypes';
import type {ReactElement} from 'react';

type TProps = {owner: string; id: string; onRerunAction: (choice: TRerunChoice, assetIDs: string[]) => void};
const stateLabels = Object.fromEntries([
 ['queued', 'Queued'], ['running', 'Running'], ['retry_wait', 'Waiting to retry'], ['blocked', 'Blocked'],
 ['succeeded', 'Succeeded'], ['failed', 'Failed'], ['canceled', 'Canceled']
]) as Record<TJobItem['state'], string>;
const outcomeLabels = {located: 'Location proposed', ambiguous: 'Ambiguous location', unknown: 'Location unknown'};

function ProgressSession({owner, id, onRerunAction}: TProps): ReactElement {
 const {data, loading: isLoading, error, stale: isStale, refresh} = useJobProgress(owner, id);
 const [isCanceling, setIsCanceling] = useState(false);
 const [cancelError, setCancelError] = useState('');
 const [selected, setSelected] = useState<string[]>([]);
 const flight = useRef<AbortController | null>(null);
 useEffect(() => () => flight.current?.abort(), []);
 const cancel = async (): Promise<void> => {
  if (flight.current) {return;}
  const controller = new AbortController(); flight.current = controller;
  setIsCanceling(true); setCancelError('');
  try {
   await cancelJob(id, controller.signal);
   if (!controller.signal.aborted) {refresh();}
  } catch {
   if (!controller.signal.aborted) {setCancelError('Cancellation could not be confirmed. Refresh progress or cancel again.');}
  } finally {
   if (!controller.signal.aborted) {flight.current = null; setIsCanceling(false);}
  }
 };
 return <section aria-label={'AI job progress'} className={'space-y-3'}>
  <p className={'break-all text-xs'}>{`Job ${id}`}</p>
  <button type={'button'} onClick={refresh}>{'Refresh progress'}</button>
  {isLoading && <p role={'status'}>{'Loading progress…'}</p>}
  {error && <p role={'alert'}>{error}</p>}
  {isStale && <p>{'Displayed progress may be stale.'}</p>}
  {data && <>
   <div role={'status'} aria-live={'polite'} aria-atomic={true}>
   <p>{data.blocked ? 'Job blocked' : data.canceled ? 'Cancellation requested' : jobIsActive(data) ? 'Job in progress' : 'Job complete'}</p>
   <p>{Object.entries(data.counts).map(([state, count]) => `${state === 'total' ? 'Total' : stateLabels[state as TJobItem['state']] ?? state}: ${count}`).join(' · ')}</p>
   </div>
   <p>{`Calls: ${data.usage.calls} · Reserved tokens: ${data.usage.reservedTokens} · Reported tokens: ${data.usage.totalReported ?? 'Unknown'} (${data.usage.reportedStatus})`}</p>
   <p>{data.usage.estimatedMicros === null ? 'Estimated cost: Unknown' : `Estimated cost: ${data.usage.estimatedMicros / 1000000} ${data.usage.currency ?? ''}`}</p>
   <p>{'Cancellation preserves completed results. It cannot recall transmitted data or usage already incurred.'}</p>
   <button type={'button'} disabled={isCanceling || data.canceled || !data.items.some(item => ['queued', 'running', 'retry_wait'].includes(item.state))} onClick={cancel}>{isCanceling ? 'Canceling…' : 'Cancel remaining work'}</button>
   {cancelError && <p role={'alert'}>{cancelError}</p>}
   <p>{'A rerun creates a separate job after a fresh preview and consent. Earlier results are retained.'}</p>
   <button type={'button'} disabled={!data.items.some(item => item.state === 'failed')} onClick={() => onRerunAction({parentJobId: id, kind: 'retry-failed'}, data.items.filter(item => item.state === 'failed').map(item => item.assetId))}>{'Preview retry of failed assets'}</button>
   <button type={'button'} disabled={selected.length === 0} onClick={() => onRerunAction({parentJobId: id, kind: 'reanalysis'}, selected)}>{'Preview reanalysis of selected assets'}</button>
   <ul className={'space-y-2'}>{data.items.map(item => <li key={item.id} className={'rounded border border-(--color-border) p-2'}>
    <label className={'flex items-center gap-2'}><input type={'checkbox'} aria-label={`Reanalyze ${item.assetId}`} checked={selected.includes(item.assetId)} onChange={event => setSelected(current => event.target.checked ? [...current, item.assetId] : current.filter(asset => asset !== item.assetId))} />{'Select for reanalysis'}</label>
    <p className={'break-all text-xs'}>{item.assetId}</p>
    <p>{`${stateLabels[item.state]}${item.outcome ? ` · ${outcomeLabels[item.outcome]}` : ''}`}</p>
    {item.failure && <p>{`Failure: ${item.failure}`}</p>}
    {item.resultId && <p className={'break-all text-xs'}>{`Saved result: ${item.resultId}`}</p>}
                                                       </li>)}
   </ul>
           </>}
        </section>;
}

export function JobProgressPanel(props: TProps): ReactElement {
 return <ProgressSession key={`${props.owner}:${props.id}`} {...props} />;
}
