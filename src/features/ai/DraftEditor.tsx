import {useEffect, useMemo, useState} from 'react';

import {cameraInput} from './draftEdit';
import {DraftMapDynamic} from './DraftMapDynamic';


import type {TDraftEdit} from './draftEdit';
import type {TDraft, TDraftPoint} from './draftTypes';
import type {ReactElement} from 'react';

export function DraftEditor({draft, isBusy, onSave, onDirtyChange, hasManualConflict = false, candidates = []}: {draft: TDraft; candidates?: {id: string; name: string; camera: TDraftPoint | null}[]; isBusy: boolean; hasManualConflict?: boolean; onDirtyChange?: (dirty: boolean) => void; onSave: (edit: TDraftEdit) => Promise<void>}): ReactElement {
 const [candidate, setCandidate] = useState(draft.candidateId || '');
 const [latitude, setLatitude] = useState(draft.camera?.latitude.toString() ?? '');
 const [longitude, setLongitude] = useState(draft.camera?.longitude.toString() ?? '');
 const [heading, setHeading] = useState(draft.heading?.toString() ?? ''); const [hasReviewedHeading, setHasReviewedHeading] = useState(false);
 const [hasGPS, setHasGPS] = useState(draft.fields.includes('gps'));
 const [descriptions, setDescriptions] = useState<NonNullable<TDraftEdit['descriptions']>>({});
 const [error, setError] = useState('');
 const point = useMemo(() => {try {return cameraInput(latitude, longitude);} catch {return null;}}, [latitude, longitude]);
 const haveCameraFactsChanged = candidate !== (draft.candidateId || '') || latitude !== (draft.camera?.latitude.toString() ?? '') || longitude !== (draft.camera?.longitude.toString() ?? '');
 const isDirty = candidate !== (draft.candidateId || '') || latitude !== (draft.camera?.latitude.toString() ?? '') || longitude !== (draft.camera?.longitude.toString() ?? '') || heading !== (draft.heading?.toString() ?? '') || hasReviewedHeading || hasGPS !== draft.fields.includes('gps') || Object.keys(descriptions).length > 0;
 useEffect(() => {onDirtyChange?.(isDirty); return () => onDirtyChange?.(false);}, [isDirty, onDirtyChange]);
 async function save(state?: TDraft['state']): Promise<void> {
  try {
   const camera = cameraInput(latitude, longitude);
   const edit: TDraftEdit = {fields: hasGPS ? ['gps'] : [], ...(candidate && candidate !== draft.candidateId ? {candidateId: candidate} : {camera})};
   if (heading !== (draft.heading?.toString() ?? '')) {
    const value = heading.trim() ? Number(heading) : null;
    if (value !== null && (!Number.isFinite(value) || value < 0 || value >= 360)) {throw new Error('Heading must be empty or from 0 up to 360 degrees, excluding 360.');}
    edit.heading = value;
   } else if (hasReviewedHeading) {edit.reviewHeading = true;}
   if (Object.keys(descriptions).length) {edit.descriptions = descriptions;}
   if (state) {edit.state = state;}
   setError(''); await onSave(edit);
  } catch (failure) {setError(failure instanceof Error ? failure.message : 'Invalid camera coordinates.');}
 }
 return <fieldset disabled={isBusy} inert={isBusy} className={'min-w-0 space-y-3'}>
  {isDirty && <p role={'status'}>{'Unsaved local edits'}</p>}
  <p>{`Exact photo: ${draft.assetId}`}</p>
  <p>{`Saved camera: ${draft.camera ? `${draft.camera.latitude}, ${draft.camera.longitude}` : 'absent'}`}</p>
  {candidates.length > 0 && <label className={'grid! gap-1 text-sm'}>{'Draft camera candidate'}<select
value={candidate} onChange={event => {
   const chosen = candidates.find(item => item.id === event.target.value); setCandidate(event.target.value);
   if (chosen) {setLatitude(chosen.camera?.latitude.toString() ?? ''); setLongitude(chosen.camera?.longitude.toString() ?? '');}
  }}><option value={''}>{'User supplied point'}</option>{candidates.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}
                                                                                               </select>
                            </label>}
  <label className={'grid! gap-1 text-sm'}>{'Camera latitude'}<input className={'w-full rounded border p-2'} type={'number'} min={-90} max={90} step={'any'} value={latitude} onChange={event => {setLatitude(event.target.value); setCandidate('');}} /></label>
  <label className={'grid! gap-1 text-sm'}>{'Camera longitude'}<input className={'w-full rounded border p-2'} type={'number'} min={-180} max={180} step={'any'} value={longitude} onChange={event => {setLongitude(event.target.value); setCandidate('');}} /></label>
  <DraftMapDynamic point={point} onPoint={next => {setCandidate(''); setLatitude(String(next.latitude)); setLongitude(String(next.longitude));}} />
  <p>{`Radius: ${draft.radius === null ? 'unknown' : `${draft.radius} m`} · ${(draft.radiusStale || (haveCameraFactsChanged && draft.radius !== null)) ? 'needs review; old estimate does not support this point' : draft.radiusBasis || 'no estimate'}`}</p>
  <p>{`Heading: ${(draft.headingStale || (haveCameraFactsChanged && draft.heading !== null)) ? 'needs review' : draft.headingUserSupplied ? 'user supplied' : 'inherited estimate'}`}</p>
  <label className={'grid! gap-1 text-sm'}>{'Heading degrees'}<input className={'w-full rounded border p-2'} type={'number'} min={0} max={359.999999} step={'any'} value={heading} onChange={event => setHeading(event.target.value)} /></label>
  <label className={'flex! items-center gap-2 text-sm'}><input className={'w-auto!'} type={'checkbox'} checked={hasReviewedHeading} onChange={event => setHasReviewedHeading(event.target.checked)} />{'Review heading for this camera point'}</label>
  {draft.descriptions.map(item => <div key={item.language}>
   <p>{`${item.language}: ${item.status} · ${(item.stale || (haveCameraFactsChanged && item.basis === 'candidate')) ? 'needs review' : item.userSupplied ? 'user supplied' : 'current'}${item.basis === 'scene_only' ? ' · scene only' : ''}`}</p>
   <label className={'grid! gap-1 text-sm'}>{`Description ${item.language}`}<textarea className={'min-h-24 w-full rounded border p-2'} maxLength={16384} value={descriptions[item.language]?.text ?? item.text ?? ''} onChange={event => setDescriptions(current => ({...current, [item.language]: {text: event.target.value}}))} /></label>
   {item.status === 'complete' && <label className={'flex! items-center gap-2 text-sm'}><input
className={'w-auto!'} type={'checkbox'} checked={!!descriptions[item.language]?.review} onChange={event => setDescriptions(current => {
    const next = {...current}; if (event.target.checked) {next[item.language] = {...next[item.language], review: true};} else if (next[item.language]?.text !== undefined) {next[item.language] = {text: next[item.language].text};} else {delete next[item.language];} return next;
   })} />{`Review ${item.language} for this camera point`}
                                  </label>}
                                  </div>)}
  <label className={'flex! items-center gap-2 text-sm'}><input className={'w-auto!'} type={'checkbox'} checked={hasGPS} onChange={event => setHasGPS(event.target.checked)} />{'Select GPS for this photo'}</label>
  {hasManualConflict && <p role={'alert'}>{'A manual location is pending for this photo. Explicitly save or discard that manual change before staging AI. Both decisions are preserved.'}</p>}
  <p>{'Only GPS can be staged. Heading and descriptions stay local.'}</p>
  {error && <p role={'alert'}>{error}</p>}
  <div className={'flex flex-wrap gap-2'}>
   <button type={'button'} disabled={isBusy} onClick={() => void save()}>{'Save draft'}</button>
   <button type={'button'} disabled={isBusy || !hasGPS || hasManualConflict} onClick={() => void save('staged')}>{'Stage GPS draft'}</button>
   <button type={'button'} disabled={isBusy} onClick={() => void save(draft.state === 'rejected' ? 'draft' : 'rejected')}>{draft.state === 'rejected' ? 'Reopen draft' : 'Reject draft'}</button>
  </div>
        </fieldset>;
}
