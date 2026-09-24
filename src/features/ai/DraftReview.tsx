import {useCallback, useEffect, useRef, useState} from 'react';

import {AIRequestError} from './aiRequest';
import {acceptDraft, fetchDraft, saveDraft} from './draftApi';
import {DraftBaseline} from './DraftBaseline';
import {DraftEditor} from './DraftEditor';

import type {TDraftEdit} from './draftEdit';
import type {TDraft} from './draftTypes';
import type {TResultDetail} from './resultDetailTypes';
import type {TReview} from './reviewTypes';
import type {ReactElement} from 'react';

export function DraftReview({owner, detail, review, onDirtyChange, hasManualConflict}: {owner: string; detail: TResultDetail; review: TReview; hasManualConflict?: boolean; onDirtyChange?: (dirty: boolean) => void}): ReactElement {
 const [editorVersion, setEditorVersion] = useState(0);
 const [isDirty, setIsDirty] = useState(false);
 const dirtyChanged = useCallback((value: boolean) => {setIsDirty(value); onDirtyChange?.(value);}, [onDirtyChange]);
 const [shouldReconcile, setNeedsReconcile] = useState(false); const [comparison, setComparison] = useState<TDraft | null>(null); const [revisionOverride, setRevisionOverride] = useState<number | null>(null);
 const [draft, setDraft] = useState<TDraft | null>(null); const [error, setError] = useState(''); const [isBusy, setIsBusy] = useState(false);
 const [candidate, setCandidate] = useState(review.initialCandidate || ''); const controller = useRef<AbortController | null>(null);
 useEffect(() => {
  const active = new AbortController(); controller.current = active;
  if (detail.entry.draftId) {
   setIsBusy(true);
   void fetchDraft(detail.entry.draftId, active.signal).then(value => {if (!active.signal.aborted) {setDraft(value);}}).catch(() => {if (!active.signal.aborted) {setError('Draft unavailable. Reload saved state.');}}).finally(() => {if (!active.signal.aborted) {setIsBusy(false);}});
  }
  return () => active.abort();
 }, [detail.entry.draftId]);
 async function accept(): Promise<void> {
  if (!detail.entry.analysisId) {return;}
  const signal = controller.current?.signal; setIsBusy(true); setError('');
  try {const value = await acceptDraft(detail.entry.analysisId, candidate || null, signal); if (!signal?.aborted) {setDraft(value);}}
  catch {if (!signal?.aborted) {setError('Acceptance not acknowledged. Retry acceptance to retrieve the same saved draft.');}}
  finally {if (!signal?.aborted) {setIsBusy(false);}}
 }
 async function save(edit: TDraftEdit): Promise<void> {
  if (!draft) {return;}
  const signal = controller.current?.signal; setIsBusy(true); setError('');
  try {const value = await saveDraft({...draft, revision: revisionOverride ?? draft.revision}, edit, signal); if (!signal?.aborted) {setDraft(value); setRevisionOverride(null);}}
  catch (failure) {if (!signal?.aborted) {setNeedsReconcile(true); setError(failure instanceof AIRequestError && failure.code === 'DRAFT_CONFLICT' ? 'Another tab saved a newer revision. Your edits are preserved.' : 'Save outcome unknown. Compare saved revision before retrying.');}}
  finally {if (!signal?.aborted) {setIsBusy(false);}}
 }
 async function compare(): Promise<void> {
  if (!draft) {return;}
  const signal = controller.current?.signal; setIsBusy(true);
  try {const value = await fetchDraft(draft.id, signal); if (!signal?.aborted) {setComparison(value);}}
  catch {if (!signal?.aborted) {setError('Saved state unavailable. Your edits remain local; try comparison again.');}}
  finally {if (!signal?.aborted) {setIsBusy(false);}}
 }
 return <section aria-label={'Local review draft'} className={'space-y-3 rounded border p-3'}>
  <h3>{'Local review decision'}</h3><p>{'Saving or staging a draft does not change Immich.'}</p>
  {error && <p role={'alert'}>{error}</p>}
  {shouldReconcile && <button type={'button'} disabled={isBusy} onClick={() => void compare()}>{'Compare saved revision'}</button>}
  {comparison && <div>
   <p>{`Saved elsewhere: revision ${comparison.revision} · Camera ${comparison.camera ? `${comparison.camera.latitude}, ${comparison.camera.longitude}` : 'absent'}`}</p>
   <p>{`State: ${comparison.state} · Heading: ${comparison.heading ?? 'unknown'}`}</p>
   {comparison.descriptions.map(item => <p key={item.language}>{`${item.language}: ${item.text ?? 'unavailable'}`}</p>)}
   <button type={'button'} onClick={() => {setDraft(comparison); setEditorVersion(value => value + 1); setComparison(null); setNeedsReconcile(false); setRevisionOverride(null); setError('');}}>{'Discard edits and use saved revision'}</button>
   <button type={'button'} onClick={() => {setRevisionOverride(comparison.revision); setComparison(null); setNeedsReconcile(false); setError('Reviewed saved state. Save explicitly to apply your local edits.');}}>{'Keep edits using saved revision'}</button>
                 </div>}
  {draft ? <><p role={'status'}>{`Saved ${draft.state} · Revision ${draft.revision}`}</p><DraftEditor key={`${draft.id}:${draft.revision}:${editorVersion}`} draft={draft} candidates={review.candidates} hasManualConflict={hasManualConflict} isBusy={isBusy || shouldReconcile} onSave={save} onDirtyChange={dirtyChanged} /><DraftBaseline key={`baseline:${draft.id}:${draft.revision}`} owner={owner} entry={detail.entry} draft={draft} disabled={isBusy || isDirty || shouldReconcile} onSaved={setDraft} onUncertain={() => setNeedsReconcile(true)} /></> : <>
   {review.candidates.length > 0 && <label>{'Candidate for draft'}<select value={candidate} onChange={event => setCandidate(event.target.value)}><option value={''}>{'Choose explicitly'}</option>{review.candidates.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>}
   <button type={'button'} disabled={isBusy || (review.outcome === 'ambiguous' && !candidate)} onClick={() => void accept()}>{'Accept as local draft'}</button>
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      </>}
        </section>;
}
