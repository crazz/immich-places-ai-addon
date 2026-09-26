import {useEffect, useRef, useState} from 'react';

import {createStackReview} from './stackReviewApi';

import type {TDraft} from './draftTypes';
import type {TStackReview} from './stackReviewTypes';
import type {ReactElement} from 'react';

type TProps = {owner: string; draft: TDraft; disabled?: boolean; hasManualConflict?: boolean; onChange: (review: TStackReview | null, ids: string[], expired?: boolean) => void};
export function StackReview({owner, draft, disabled, hasManualConflict, onChange}: TProps): ReactElement {
 const [review, setReview] = useState<TStackReview | null>(null);
 const [selected, setSelected] = useState([draft.assetId]);
 const [acknowledged, setAcknowledged] = useState<string[]>([]);
 const [isObservationCurrent, setObservationCurrent] = useState(false);
 const [hasExpired, setExpired] = useState(false);
 const [isBusy, setBusy] = useState(false); const [error, setError] = useState('');
 const active = useRef<AbortController | null>(null);
 useEffect(() => () => active.current?.abort(), []);
 useEffect(() => {
  if (!review) {return;}
  const remaining = Date.parse(review.expiresAt) - Date.now();
  const expire = (): void => {setExpired(true); setAcknowledged([]); onChange(null, selected, true);};
  if (remaining <= 0) {expire(); return;}
  const timer = setTimeout(expire, remaining); return () => clearTimeout(timer);
 }, [review, selected, onChange]);
 useEffect(() => {
  if ((disabled || hasManualConflict) && (review || active.current)) {active.current?.abort(); setBusy(false); setObservationCurrent(false); setAcknowledged([]); onChange(null, selected);}
 }, [disabled, hasManualConflict, onChange, selected, review]);
 async function read(ids: string[]): Promise<void> {
  active.current?.abort(); const controller = new AbortController(); active.current = controller;
  setExpired(false); setObservationCurrent(false); setBusy(true); setError(''); setAcknowledged([]); onChange(null, ids);
  try {const value = await createStackReview(owner, draft, ids, controller.signal); if (!controller.signal.aborted) {setReview(value); setObservationCurrent(true);}}
  catch {if (!controller.signal.aborted) {setError('Selected stack sources or baselines unavailable. Keep the selection and explicitly read it again.');}}
  finally {if (!controller.signal.aborted) {setBusy(false);}}
 }
 function select(id: string, include: boolean): void {
  const ids = include ? [...selected, id].sort() : selected.filter(item => item !== id);
  setSelected(ids); setAcknowledged([]); onChange(null, ids);
 }
 function acknowledge(id: string, checked: boolean): void {
  const ids = checked ? [...acknowledged, id] : acknowledged.filter(item => item !== id);
  setAcknowledged(ids); onChange(review && ids.length === selected.length ? review : null, selected);
 }
 const isExact = isObservationCurrent && !hasExpired && review?.targets.length === selected.length && review.targets.every(target => selected.includes(target.assetId));
 return <section aria-label={'Stack GPS targets'} className={'space-y-2'}>
  <p>{'A stack may contain different camera viewpoints. Review each source and GPS baseline independently.'}</p>
  <button type={'button'} disabled={disabled || isBusy} onClick={() => void read(selected)}>{'Review stack members'}</button>
  <p>{`${selected.length} selected · Maximum 50 including the analyzed photo.`}</p>
  {review && <>
   {review.candidates.map(id => <label key={id} className={'block break-all'}><input type={'checkbox'} aria-label={`Include photo ${id}`} checked={selected.includes(id)} disabled={disabled || isBusy || id === draft.assetId || (selected.length >= 50 && !selected.includes(id))} onChange={event => select(id, event.target.checked)} />{`Photo ${id}${id === draft.assetId ? ' (analyzed; always included)' : ''}`}</label>)}
   <button type={'button'} disabled={disabled || hasManualConflict || isBusy} onClick={() => void read(selected)}>{'Read selected target baselines'}</button>
   {isExact && review.targets.map(target => <div key={target.assetId} className={'break-all rounded border p-2'}>
    <p>{`Photo ${target.assetId} · Source available: ${target.imageIdentity}`}</p>
    <p>{`Before GPS: ${target.before.latitude ?? 'absent'}, ${target.before.longitude ?? 'absent'}`}</p>
    <p>{`Proposed GPS: ${draft.camera?.latitude}, ${draft.camera?.longitude}`}</p>
    <label><input type={'checkbox'} aria-label={`Reviewed source and GPS for ${target.assetId}`} checked={acknowledged.includes(target.assetId)} disabled={disabled || hasManualConflict || isBusy} onChange={event => acknowledge(target.assetId, event.target.checked)} />{'I reviewed this source and GPS baseline.'}</label>
                                            </div>)}
   <p>{`Target review expires: ${review.expiresAt}`}</p>
             </>}
  {hasExpired && <p role={'alert'}>{'Target review expired. Read and acknowledge the selected baselines again.'}</p>}
  {error && <p role={'alert'}>{error}</p>}
        </section>;
}
