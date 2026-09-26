import {useCallback, useEffect, useRef, useState} from 'react';

import {isDraftPoint} from './draftTypes';
import {MirrorComparison} from './MirrorComparison';
import {StackReview} from './StackReview';
import {StackWriteComparison} from './StackWriteComparison';
import {StandardWriteComparison} from './StandardWriteComparison';
import {PreviewRequestError, createWritePreview, fetchWritePreview} from './writePreviewApi';

import type {TDraft} from './draftTypes';
import type {TStackReview} from './stackReviewTypes';
import type {TPreviewConflict, TWritePreview} from './writePreviewTypes';
import type {ReactElement} from 'react';

type TProps = {owner: string; draft: TDraft; disabled?: boolean; hasManualConflict?: boolean; manualPendingIDs?: string[]; onTargetsChange?: (ids: string[]) => void; onCompare?: () => void; onPreview?: (preview: TWritePreview | null) => void};

export function WritePreview(props: TProps): ReactElement {
 return <PreviewSession key={`${props.owner}:${props.draft.id}:${props.draft.revision}`} {...props} hasManualConflict={props.hasManualConflict && props.draft.fields.includes('gps')} />;
}

function PreviewSession({owner, draft, disabled, hasManualConflict: hasPrimaryManualConflict, manualPendingIDs = [], onTargetsChange, onCompare, onPreview}: TProps): ReactElement {
 const [hasMirrorDisclosure, setMirrorDisclosure] = useState(false);
 const [stackReview, setStackReview] = useState<TStackReview | null>(null);
 const [targetIds, setTargetIds] = useState([draft.assetId]);
 const hasDescription = draft.fields.includes('description');
 const [preview, setPreview] = useState<TWritePreview | null>(null);
 const selectedIds = (preview?.plan.version === 'stack-preview-v3' || preview?.plan.version === 'mirror-preview-v4') ? preview.plan.manifest.targets.map(target => target.assetId) : targetIds;
 const conflicts = draft.fields.includes('gps') ? selectedIds.filter(id => manualPendingIDs.includes(id)) : [];
 const hasManualConflict = hasPrimaryManualConflict || conflicts.length > 0;
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
   void fetchWritePreview(owner, draft, id, controller.signal).then(value => {if (!controller.signal.aborted) {setPreview(value); if ((value.plan.version === 'stack-preview-v3' || value.plan.version === 'mirror-preview-v4')) {const ids = value.plan.manifest.targets.map(target => target.assetId); setTargetIds(ids); onTargetsChange?.(ids);}}}).catch(() => {if (!controller.signal.aborted) {setError('Stored preview unavailable. Create a fresh comparison.');}}).finally(() => {if (!controller.signal.aborted) {setIsBusy(false);}});
  }
  return () => {active.current?.abort(); try {sessionStorage.removeItem(storageKey);} catch { /* No persisted reference to clear. */ }};
 }, [owner, draft, storageKey, onTargetsChange]);
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
 const stackChanged = useCallback((review: TStackReview | null, ids: string[], expired = false): void => {
  if (expired) {setStackReview(null); return;}
  active.current?.abort(); active.current = new AbortController(); setIsBusy(false); setPreview(null); setStackReview(review); setTargetIds(ids); onTargetsChange?.(ids);
  try {sessionStorage.removeItem(storageKey);} catch { /* Browser storage can be disabled. */ }
 }, [storageKey, onTargetsChange]);
 async function create(): Promise<void> {
  const signal = active.current?.signal; setIsBusy(true); setError(''); setConflict(null); setPreview(null); setShouldCompare(false);
  try {sessionStorage.removeItem(storageKey);} catch { /* Browser storage can be disabled. */ }
  try {const value = await createWritePreview(owner, draft, signal, stackReview ?? undefined, hasMirrorDisclosure ? 'asset-readers-v1' : undefined); if (!signal?.aborted) {setPreview(value); try {sessionStorage.setItem(storageKey, value.plan.id);} catch { /* The comparison remains usable without browser storage. */ }}}
  catch (failure) {if (!signal?.aborted) {
   setError(failure instanceof PreviewRequestError ? previewRecovery(failure.code, hasDescription) : 'Preview unavailable. Review the saved draft and current source, then retry.');
   if (failure instanceof PreviewRequestError && failure.conflict) {setConflict(failure.conflict);}
   setShouldCompare(failure instanceof PreviewRequestError && ['DRAFT_CONFLICT', 'DRAFT_NOT_STAGED'].includes(failure.code));
  }}
  finally {if (!signal?.aborted) {setIsBusy(false);}}
 }
 return <section aria-label={hasDescription ? 'Exact selected fields preview' : 'Exact GPS preview'} className={'space-y-2 border-t pt-3'}>
  <h4 className={'font-semibold'}>{hasDescription ? 'Exact selected fields comparison' : 'Exact GPS comparison'}</h4>
  <p>{draft.mirror ? 'Selected standard fields and the analyzed photo’s optional metadata will be compared independently.' : targetIds.length > 1 ? 'Explicit stack targets · GPS for every selected photo; description only for the analyzed photo.' : hasDescription ? 'Single photo · Only the selected fields will be written.' : 'Single photo · GPS only. Other metadata is preserved.'}</p>
  <p>{'Creating this comparison does not change Immich. Source read access does not verify write permission.'}</p>
  <StackReview hasManualConflict={hasManualConflict} owner={owner} draft={draft} disabled={disabled || draft.state !== 'staged' || !draft.fields.includes('gps') || !isDraftPoint(draft.camera)} onChange={stackChanged} />
  {draft.mirror && <label><input type={'checkbox'} checked={hasMirrorDisclosure} disabled={disabled || isBusy} onChange={event => {setMirrorDisclosure(event.target.checked); if (!event.target.checked) {active.current?.abort(); active.current = new AbortController(); setPreview(null); try {sessionStorage.removeItem(storageKey);} catch { /* No retained preview to clear. */ }}}} />{'I understand authorized asset readers may see this exported information'}</label>}
  <button type={'button'} disabled={disabled || (draft.mirror && !hasMirrorDisclosure) || hasManualConflict || isBusy || (targetIds.length > 1 && !stackReview) || draft.state !== 'staged' || draft.baseline.status !== 'reviewed'} onClick={() => void create()}>{hasDescription ? 'Preview selected fields' : 'Preview exact GPS'}</button>
  {conflicts.map(id => <p key={id} role={'alert'}>{`Resolve pending manual GPS for ${id} by explicitly saving or discarding that manual change.`}</p>)}
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
   <p>{`${hasDescription ? 'Selected fields' : 'GPS'} comparison: ${preview.diff}`}</p><p role={'status'}>{`Preview status: ${hasExpired ? 'expired' : preview.status}`}</p>
   {(hasExpired || preview.status !== 'usable') && <p>{'This preview is unusable. Keep the saved draft and create a fresh comparison after review.'}</p>}
   {(preview.plan.version === 'stack-preview-v3' || preview.plan.version === 'mirror-preview-v4') ? <StackWriteComparison plan={preview.plan} /> : <StandardWriteComparison plan={preview.plan} />}
   {preview.plan.version === 'mirror-preview-v4' && <MirrorComparison mirror={preview.plan.mirror} targetId={preview.plan.targetId} />}
   <p>{`Expires: ${preview.plan.expiresAt}`}</p>
   <p className={'break-all text-xs'}>{`Plan digest: ${preview.digest}`}</p>
                                                 </>}
        </section>;
}

function previewRecovery(code: string, hasDescription: boolean): string {
 const sourceReview = hasDescription ? 'Review current source and selected fields' : 'Review current source and GPS';
 switch (code) {
  case 'INVALID_MIRROR': return 'Selected metadata needs current review. Review or deselect stale contents, save and stage a new revision.';
  case 'MIRROR_TOO_LARGE': return 'Selected metadata exceeds the supported size. Choose fewer languages or shorten reviewed text, then stage a new revision.';
  case 'METADATA_UNAVAILABLE': return 'Current metadata is unavailable. Your draft is preserved; retry the comparison when access returns.';
  case 'METADATA_CONFLICT': return 'Owned metadata changed. Create a fresh comparison and review the full existing export before confirming.';
  case 'METADATA_UNSUPPORTED': return 'Metadata mirroring is unsupported or unverified. Local records are preserved. Deselect mirroring and stage a new revision to continue with standard fields only.';
  case 'DESCRIPTION_INVALID': return 'Description is not ready. Select complete current text, review or regenerate it, and save a new revision.';
  case 'DESCRIPTION_TOO_LARGE': return 'Description exceeds the supported size. Edit the selected text or choose replacement, then review a new comparison.';
  case 'POLICY_CHANGED': return 'Writing policy changed. Keep the saved draft and create a fresh comparison after reviewing the current writing capability.';
  case 'DESCRIPTION_CONFLICT': return 'Description changed. Use Review current source and selected fields, explicitly acknowledge the current text, then stage a new revision.';
  case 'IMMICH_CONFLICT': return 'GPS changed. Use Review current source and GPS, explicitly acknowledge it, then stage a new revision.';
  case 'SOURCE_CHANGED': return `The source image changed. Use ${sourceReview} to inspect the current image, then acknowledge and restage.`;
  case 'BASELINE_REVIEW_REQUIRED': return `Use ${sourceReview} to acknowledge a baseline, then stage a new revision.`;
  case 'DRAFT_CONFLICT': return 'The saved draft revision changed. Compare saved revision before creating another preview.';
  case 'DRAFT_NOT_STAGED': return 'The saved draft is no longer staged. Reload saved state and stage the reviewed GPS decision.';
  case 'PREVIEW_CAPACITY': return 'Ten active previews already exist. Wait for their five-minute validity to end before creating another.';
  case 'SOURCE_UNAVAILABLE': return 'Current source temporarily unavailable. Your saved draft is preserved; retry the preview when access returns.';
  default: return 'Preview unavailable. Review the saved draft and current source, then retry.';
 }
}
