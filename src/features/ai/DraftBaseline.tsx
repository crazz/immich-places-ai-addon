import {useEffect, useRef, useState} from 'react';

import {acknowledgeDraft, observeDraft} from './draftBaselineApi';
import {ResultImage} from './ResultImage';

import type {TDraftObservation} from './draftBaselineApi';
import type {TDraft} from './draftTypes';
import type {TResultEntry} from './resultTypes';
import type {ReactElement} from 'react';

export function DraftBaseline({owner, entry, draft, disabled, onSaved, onUncertain}: {owner: string; entry: TResultEntry; draft: TDraft; disabled: boolean; onSaved: (draft: TDraft) => void; onUncertain: () => void}): ReactElement {
 const [observation, setObservation] = useState<TDraftObservation | null>(null); const [error, setError] = useState('');
 const [isBusy, setIsBusy] = useState(false); const [hasImage, setHasImage] = useState(false); const [hasReviewed, setHasReviewed] = useState(false);
 const active = useRef<AbortController | null>(null);
 useEffect(() => {active.current = new AbortController(); return () => active.current?.abort();}, []);
 async function observe(): Promise<void> {
  const signal = active.current?.signal; setIsBusy(true); setError(''); setObservation(null); setHasReviewed(false);
  try {
   const value = await observeDraft(draft, signal);
   if (value.draftId !== draft.id || value.revision !== draft.revision || value.baseline.assetId !== draft.assetId) {throw new Error('Invalid observation');}
   if (!signal?.aborted) {setObservation(value);}
  } catch {if (!signal?.aborted) {setError('Current source unavailable or changed. Local edits remain saved. Review again when access is available.');}}
  finally {if (!signal?.aborted) {setIsBusy(false);}}
 }
 async function acknowledge(): Promise<void> {
  if (!observation || !hasImage || !hasReviewed) {return;}
  const signal = active.current?.signal; setIsBusy(true); setError('');
  try {const value = await acknowledgeDraft(draft, observation, signal); if (!signal?.aborted) {onSaved(value);}}
  catch {if (!signal?.aborted) {setObservation(null); setError('Baseline acknowledgement uncertain or obsolete. Compare saved state, then renew source review.'); onUncertain();}}
  finally {if (!signal?.aborted) {setIsBusy(false);}}
 }
 return <section aria-label={'Current GPS baseline'} className={'space-y-2 border-t pt-3'}>
  <p>{`Baseline: ${draft.baseline.status}`}</p>
  {draft.baseline.status === 'reviewed' && <p>{`Reviewed GPS: latitude ${draft.baseline.latitude ?? 'absent'} · longitude ${draft.baseline.longitude ?? 'absent'}`}</p>}
  <button type={'button'} disabled={disabled || isBusy} onClick={() => void observe()}>{'Review current source and GPS'}</button>
  {error && <p role={'alert'}>{error}</p>}
  {observation && <>
   <ResultImage key={observation.id} owner={owner} entry={{...entry, sourceAvailable: true}} onAvailable={setHasImage} />
   <p>{`Observed GPS: latitude ${observation.baseline.latitude ?? 'absent'} · longitude ${observation.baseline.longitude ?? 'absent'}`}</p>
   <p>{observation.originalSourceMatches ? 'Original source fingerprint matches this observation.' : 'Original source fingerprint differs. Review the current image; the old analysis remains historical.'}</p>
   <p className={'break-all'}>{`Reviewed image identity: ${observation.baseline.imageIdentity}`}</p>
   <p>{`Observation expires: ${observation.expiresAt}`}</p>
   <label className={'flex! items-center gap-2'}><input className={'w-auto!'} type={'checkbox'} checked={hasReviewed} disabled={!hasImage || disabled || isBusy} onChange={event => setHasReviewed(event.target.checked)} />{'I reviewed this current image and its displayed GPS'}</label>
   <button type={'button'} disabled={!hasImage || !hasReviewed || disabled || isBusy} onClick={() => void acknowledge()}>{'Acknowledge displayed baseline'}</button>
                  </>}
        </section>;
}
