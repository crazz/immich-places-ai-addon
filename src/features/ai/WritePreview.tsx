import {useEffect, useRef, useState} from 'react';

import {PreviewRequestError, createWritePreview, fetchWritePreview} from './writePreviewApi';

import type {TDraft} from './draftTypes';
import type {TPreviewConflict, TWritePreview} from './writePreviewTypes';
import type {ReactElement} from 'react';

type TProps = {owner: string; draft: TDraft; disabled?: boolean; hasManualConflict?: boolean; onCompare?: () => void; onPreview?: (preview: TWritePreview | null) => void};

export function WritePreview(props: TProps): ReactElement {
 return <PreviewSession key={`${props.owner}:${props.draft.id}:${props.draft.revision}`} {...props} />;
}

function PreviewSession({owner, draft, disabled, hasManualConflict, onCompare, onPreview}: TProps): ReactElement {
 const [preview, setPreview] = useState<TWritePreview | null>(null);
 const [error, setError] = useState('');
 const [conflict, setConflict] = useState<TPreviewConflict | null>(null);
 const [shouldCompare, setShouldCompare] = useState(false);
 const [isBusy, setIsBusy] = useState(false);
 const [hasExpired, setHasExpired] = useState(false);
 const active = useRef<AbortController | null>(null);
 const storageKey = `ai-write-preview:${owner}:${draft.id}:${draft.revision}`;
 useEffect(() => {
  onPreview?.(preview?.status === 'usable' && !hasExpired && !disabled && !hasManualConflict ? preview : null);
  return () => onPreview?.(null);
 }, [preview, hasExpired, disabled, hasManualConflict, onPreview]);
 useEffect(() => {
  const controller = new AbortController(); active.current = controller;
  let id: string | null = null;
  try {id = sessionStorage.getItem(storageKey);} catch { /* Browser storage can be disabled. */ }
  if (id) {
   setIsBusy(true);
   void fetchWritePreview(owner, draft, id, controller.signal).then(value => {if (!controller.signal.aborted) {setPreview(value);}}).catch(() => {if (!controller.signal.aborted) {setError('Stored preview unavailable. Create a fresh comparison.');}}).finally(() => {if (!controller.signal.aborted) {setIsBusy(false);}});
  }
  return () => {active.current?.abort(); try {sessionStorage.removeItem(storageKey);} catch { /* No persisted reference to clear. */ }};
 }, [owner, draft, storageKey]);
 useEffect(() => {
  if (disabled || hasManualConflict) {
   active.current?.abort(); active.current = new AbortController(); setPreview(null); setIsBusy(false);
   try {sessionStorage.removeItem(storageKey);} catch { /* Browser storage can be disabled. */ }
  }
 }, [disabled, hasManualConflict, storageKey]);
 useEffect(() => {
  const remaining = preview ? Date.parse(preview.plan.expiresAt) - Date.now() : 0;
  setHasExpired(!!preview && remaining <= 0);
  if (!preview || remaining <= 0) {return;}
  const timer = setTimeout(() => setHasExpired(true), Math.min(remaining, 300_000));
  return () => clearTimeout(timer);
 }, [preview]);
 async function create(): Promise<void> {
  const signal = active.current?.signal; setIsBusy(true); setError(''); setConflict(null); setPreview(null); setShouldCompare(false);
  try {sessionStorage.removeItem(storageKey);} catch { /* Browser storage can be disabled. */ }
  try {const value = await createWritePreview(owner, draft, signal); if (!signal?.aborted) {setPreview(value); try {sessionStorage.setItem(storageKey, value.plan.id);} catch { /* The comparison remains usable without browser storage. */ }}}
  catch (failure) {if (!signal?.aborted) {
   setError(failure instanceof PreviewRequestError ? previewRecovery(failure.code) : 'Preview unavailable. Review the saved draft and current source, then retry.');
   if (failure instanceof PreviewRequestError && failure.conflict) {setConflict(failure.conflict);}
   setShouldCompare(failure instanceof PreviewRequestError && ['DRAFT_CONFLICT', 'DRAFT_NOT_STAGED'].includes(failure.code));
  }}
  finally {if (!signal?.aborted) {setIsBusy(false);}}
 }
 return <section aria-label={'Exact GPS preview'} className={'space-y-2 border-t pt-3'}>
  <h4 className={'font-semibold'}>{'Exact GPS comparison'}</h4>
  <p>{'Single photo · GPS only. Other metadata is preserved.'}</p>
  <p>{'Creating this comparison does not change Immich. Source read access does not verify write permission.'}</p>
  <button type={'button'} disabled={disabled || hasManualConflict || isBusy || draft.state !== 'staged' || draft.baseline.status !== 'reviewed'} onClick={() => void create()}>{'Preview exact GPS'}</button>
  {hasManualConflict && <p role={'alert'}>{'Resolve the pending manual location by explicitly saving or discarding it before creating a fresh AI preview. Both decisions are preserved.'}</p>}
  {disabled && <p>{'Save or reconcile local edits before creating a fresh comparison.'}</p>}
  {error && <p role={'alert'}>{error}</p>}
  {shouldCompare && onCompare && <button type={'button'} disabled={isBusy} onClick={() => {setShouldCompare(false); setError(''); onCompare();}}>{'Compare saved revision'}</button>}
  {conflict && <div>
   <p>{`Reviewed baseline: latitude ${conflict.before.latitude ?? 'absent'} · longitude ${conflict.before.longitude ?? 'absent'}`}</p>
   <p>{`Current GPS: latitude ${conflict.current.latitude ?? 'absent'} · longitude ${conflict.current.longitude ?? 'absent'}`}</p>
   <p>{`Proposed GPS: latitude ${conflict.proposed.latitude} · longitude ${conflict.proposed.longitude}`}</p>
               </div>}
  {preview && !disabled && !hasManualConflict && <>
   <p className={'break-all'}>{`Photo: ${preview.plan.targetId} · Draft revision ${preview.plan.draftRevision}`}</p>
   <p>{`GPS comparison: ${preview.diff}`}</p><p role={'status'}>{`Preview status: ${hasExpired ? 'expired' : preview.status}`}</p>
   {(hasExpired || preview.status !== 'usable') && <p>{'This preview is unusable. Keep the saved draft and create a fresh comparison after review.'}</p>}
   {([['Before latitude', preview.plan.before.latitude], ['Before longitude', preview.plan.before.longitude], ['Proposed latitude', preview.plan.intended.latitude], ['Proposed longitude', preview.plan.intended.longitude]] as const).map(([label, value]) => <label key={label} className={'grid! gap-1 text-sm'}>{label}<input readOnly value={value ?? 'absent'} className={'w-full rounded border p-2'} /></label>)}
   <p>{`Expires: ${preview.plan.expiresAt}`}</p>
   <p className={'break-all text-xs'}>{`Plan digest: ${preview.digest}`}</p>
                                                 </>}
        </section>;
}

function previewRecovery(code: string): string {
 switch (code) {
  case 'IMMICH_CONFLICT': return 'GPS changed. Use Review current source and GPS, explicitly acknowledge it, then stage a new revision.';
  case 'SOURCE_CHANGED': return 'The source image changed. Use Review current source and GPS to inspect the current image, then acknowledge and restage.';
  case 'BASELINE_REVIEW_REQUIRED': return 'Use Review current source and GPS to acknowledge a baseline, then stage a new revision.';
  case 'DRAFT_CONFLICT': return 'The saved draft revision changed. Compare saved revision before creating another preview.';
  case 'DRAFT_NOT_STAGED': return 'The saved draft is no longer staged. Reload saved state and stage the reviewed GPS decision.';
  case 'PREVIEW_CAPACITY': return 'Ten active previews already exist. Wait for their five-minute validity to end before creating another.';
  case 'SOURCE_UNAVAILABLE': return 'Current source temporarily unavailable. Your saved draft is preserved; retry the preview when access returns.';
  default: return 'Preview unavailable. Review the saved draft and current source, then retry.';
 }
}
