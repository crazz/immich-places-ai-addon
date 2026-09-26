import {useCallback, useEffect, useRef, useState} from 'react';

import {isRecord} from '@/utils/typeGuards';

import {AIRequestError} from './aiRequest';
import {mirrorOperationRequest} from './mirrorApi';
import {MirrorOutcome} from './MirrorOutcome';
import {isUUID} from './selectionTypes';
import {StackWriteOutcome} from './StackWriteOutcome';
import {StandardWriteComparison} from './StandardWriteComparison';
import {StandardWriteOutcome} from './StandardWriteOutcome';
import {fetchWriteHistory, writeOperationRequest} from './writeOperationApi';
import {isWriteConfirmation} from './writeOperationTypes';

import type {TDraft} from './draftTypes';
import type {TWriteConfirmation, TWriteHistory, TWriteOperation} from './writeOperationTypes';
import type {TWritePreview} from './writePreviewTypes';
import type {ReactElement} from 'react';

type TTargetRetry = {operationId: string; assetId: string; generation: number};

type TProps = {owner: string; draft: TDraft; preview: TWritePreview | null; disabled?: boolean; metadataDisabled?: boolean; manualPendingIDs?: string[]; onVerified?: () => void};

export function WriteOperation(props: TProps): ReactElement {
 return <OperationSession key={`${props.owner}:${props.draft.id}`} {...props} />;
}

function OperationSession({owner, draft, preview, disabled, metadataDisabled = disabled, manualPendingIDs = [], onVerified}: TProps): ReactElement {
 const mirrorStorageKey = `ai-metadata-retry:${owner}:${draft.id}`;
 const [pendingMirrorRetry, setMirrorRetry] = useState<TTargetRetry | null>(() => {
  try {const value: unknown = JSON.parse(sessionStorage.getItem(mirrorStorageKey) || 'null'); if (isRecord(value) && isUUID(value.operationId) && value.assetId === draft.assetId && Number.isSafeInteger(value.generation) && Number(value.generation) >= 1) {return {operationId: value.operationId, assetId: draft.assetId, generation: Number(value.generation)};}} catch { /* No retained metadata retry. */ }
  return null;
 });
 const targetStorageKey = `ai-target-retry:${owner}:${draft.id}`;
 const [pendingTargetRetry, setTargetRetry] = useState<TTargetRetry | null>(() => {
  try {const value: unknown = JSON.parse(sessionStorage.getItem(targetStorageKey) || 'null'); if (isRecord(value) && isUUID(value.operationId) && isUUID(value.assetId) && Number.isSafeInteger(value.generation) && Number(value.generation) >= 1) {return {operationId: value.operationId, assetId: value.assetId, generation: Number(value.generation)};}} catch { /* No retained target retry. */ }
  return null;
 });
 const hasDescription = draft.fields.includes('description');
 const [operation, setOperation] = useState<TWriteOperation | null>(null);
 const [pending, setPending] = useState<TWriteConfirmation | null>(null);
 const [history, setHistory] = useState<TWriteHistory>({items: [], nextCursor: ''});
 const [historyCursor, setHistoryCursor] = useState('');
 const [error, setError] = useState(''); const [isBusy, setBusy] = useState(false);
 const [pollEpoch, setPollEpoch] = useState(0);
 const operationID = operation?.id; const shouldPoll = !!operation && hasActiveTargets(operation);
 const active = useRef<AbortController | null>(null); const locked = useRef(false);
 const confirmationVersion = useRef(0);
 const storageKey = `ai-write-confirmation:${owner}:${draft.id}`;
 const notified = useRef(new Set<string>());
 useEffect(() => {
  if (!operation) {return;}
  const refreshKey = `ai-stack-refreshed:${owner}:${draft.id}`;
  if ((operation.plan.version === 'stack-preview-v3' || operation.plan.version === 'mirror-preview-v4')) {try {const retained: unknown = JSON.parse(sessionStorage.getItem(refreshKey) || '[]'); if (Array.isArray(retained)) {for (const id of retained) {if (typeof id === 'string') {notified.current.add(id);}}}} catch { /* Refresh deduplication remains available for this view. */ }}
  const ids = (operation.plan.version === 'stack-preview-v3' || operation.plan.version === 'mirror-preview-v4') ? operation.targets?.filter(target => target.refreshed && target.fields.some(field => field.field === 'gps' && field.status === 'verified')).map(target => `${operation.id}:${target.assetId}`) ?? [] : operation.refreshed && (operation.verified || operation.fields?.some(item => item.field === 'gps' && item.status === 'verified')) ? [operation.id] : [];
  for (const id of ids) {if (!notified.current.has(id)) {notified.current.add(id); onVerified?.();}}
  if ((operation.plan.version === 'stack-preview-v3' || operation.plan.version === 'mirror-preview-v4')) {try {sessionStorage.setItem(refreshKey, JSON.stringify([...notified.current]));} catch { /* Refresh deduplication remains available for this view. */ }}
 }, [operation, onVerified, owner, draft.id]);
 const accept = useCallback((value: TWriteOperation, confirmation?: TWriteConfirmation): void => {
  if (confirmation && (value.plan.id !== confirmation.previewId || value.digest !== confirmation.digest)) {throw new Error('Confirmation changed');}
  setOperation(value); setPending(null); setError('');
  try {sessionStorage.removeItem(storageKey);} catch { /* Server history remains available. */ }
 }, [storageKey]);
 useEffect(() => {
  const controller = new AbortController(); active.current = controller;
  const version = confirmationVersion.current;
  let retained: TWriteConfirmation | null = null;
  try {const value: unknown = JSON.parse(sessionStorage.getItem(storageKey) || 'null'); if (isWriteConfirmation(value)) {retained = value; setPending(value);}} catch { /* No usable retained confirmation. */ }
  void (async () => {
   try {
    const page = await fetchWriteHistory(draft.id, '', controller.signal);
    if (controller.signal.aborted) {return;} setHistory(page);
    if (version !== confirmationVersion.current) {return;}
    const latest = [...page.items].sort((a, b) => Date.parse(b.approvedAt) - Date.parse(a.approvedAt))[0];
    if (retained || latest) {
     const value = await writeOperationRequest(owner, draft, retained ? `/by-key/${retained.idempotencyKey}` : `/${latest.id}`, undefined, controller.signal);
     if (!controller.signal.aborted && version === confirmationVersion.current) {accept(value, retained ?? undefined);}
    }
   } catch {if (!controller.signal.aborted && version === confirmationVersion.current) {setError('Saved GPS history unavailable. Reconcile any pending confirmation before continuing.');}}
  })();
  return () => controller.abort();
 }, [owner, draft, storageKey, accept]);
 useEffect(() => {
  if (!operationID || !shouldPoll) {return;}
  const id = operationID; const controller = new AbortController(); let reads = 0; let timer: ReturnType<typeof setTimeout>;
  const poll = async (): Promise<void> => {
   try {
    const value = await writeOperationRequest(owner, draft, `/${id}`, undefined, controller.signal);
    if (controller.signal.aborted || value.id !== id) {return;} setOperation(value); reads++;
    if (hasActiveTargets(value) && reads < 30) {timer = setTimeout(() => void poll(), 2_000);}
   } catch {if (!controller.signal.aborted) {setError('Status read unavailable. Use Check status; no additional GPS mutation was requested.');}}
  };
  timer = setTimeout(() => void poll(), 2_000);
  return () => {controller.abort(); clearTimeout(timer);};
 }, [owner, draft, operationID, shouldPoll, pollEpoch]);
 function handleConfirmationRejection(failure: unknown): boolean {
  if (!(failure instanceof AIRequestError) || !['WRITE_DISABLED', 'PREVIEW_EXPIRED', 'PREVIEW_CONSUMED', 'DRAFT_CONFLICT', 'PLAN_CONFLICT', 'TARGET_BUSY', 'INVALID_WRITE', 'IDEMPOTENCY_CONFLICT'].includes(failure.code)) {return false;}
  setPending(null); try {sessionStorage.removeItem(storageKey);} catch { /* A definitive rejection grants no authority. */ }
  setError(failure.code === 'WRITE_DISABLED' ? `${hasDescription ? 'Selected fields' : 'GPS'} writing is disabled. Saved review and status remain available.` : 'Confirmation was rejected. Inspect saved GPS history and create a fresh comparison after resolving the conflict.');
  return true;
 }
 async function confirm(): Promise<void> {
  if (!preview || disabled || pending || locked.current || preview.plan.draftRevision !== draft.revision || Date.parse(preview.plan.expiresAt) <= Date.now()) {return;}
  locked.current = true; setBusy(true); setError('');
  confirmationVersion.current++;
  const signal = active.current?.signal;
  try {
   const input = {previewId: preview.plan.id, digest: preview.digest, idempotencyKey: Array.from(crypto.getRandomValues(new Uint8Array(16)), byte => byte.toString(16).padStart(2, '0')).join('')};
   sessionStorage.setItem(storageKey, JSON.stringify(input)); setPending(input);
   try {const value = await writeOperationRequest(owner, draft, '', input, signal); if (!signal?.aborted) {accept(value, input);}}
   catch (failure) {if (!signal?.aborted && !handleConfirmationRejection(failure)) {setError('Confirmation acknowledgement unavailable. Reconcile this confirmation before another approval.');}}
  } catch {if (!signal?.aborted) {setError('Confirmation could not be retained safely in this browser. No request was sent.');}}
  finally {locked.current = false; if (!signal?.aborted) {setBusy(false);}}
 }
 async function reconcile(): Promise<void> {
  if (!pending || locked.current) {return;} locked.current = true; setBusy(true); const signal = active.current?.signal;
  try {const value = await writeOperationRequest(owner, draft, `/by-key/${pending.idempotencyKey}`, undefined, signal); if (!signal?.aborted) {accept(value, pending);}}
  catch {if (!signal?.aborted) {setError('Confirmation remains unresolved. No new write was sent; check again when service access returns.');}}
  finally {locked.current = false; if (!signal?.aborted) {setBusy(false);}}
 }
 async function repeatConfirmation(): Promise<void> {
  if (!pending || disabled || locked.current || preview?.plan.id !== pending.previewId || preview.digest !== pending.digest) {return;}
  locked.current = true; setBusy(true); const signal = active.current?.signal;
  try {const value = await writeOperationRequest(owner, draft, '', pending, signal); if (!signal?.aborted) {accept(value, pending);}}
  catch (failure) {if (!signal?.aborted && !handleConfirmationRejection(failure)) {setError('Confirmation remains unresolved. The same identity was used; reconcile before another approval.');}}
  finally {locked.current = false; if (!signal?.aborted) {setBusy(false);}}
 }
 async function historyPage(cursor: string): Promise<void> {
  if (locked.current) {return;} locked.current = true; setBusy(true); const signal = active.current?.signal;
  try {const page = await fetchWriteHistory(draft.id, cursor, signal); if (!signal?.aborted) {setHistory(page); setHistoryCursor(cursor); setError('');}}
  catch {if (!signal?.aborted) {setError('Saved GPS history unavailable. Your inspected operation remains available.');}}
  finally {locked.current = false; if (!signal?.aborted) {setBusy(false);}}
 }
 async function inspect(id: string): Promise<void> {
  if (locked.current) {return;} locked.current = true; confirmationVersion.current++;
  const signal = active.current?.signal; setBusy(true);
  try {const value = await writeOperationRequest(owner, draft, `/${id}`, undefined, signal); if (!signal?.aborted && value.id === id) {setOperation(value);}}
  catch {if (!signal?.aborted) {setError('Saved GPS operation unavailable.');}}
  finally {locked.current = false; if (!signal?.aborted) {setBusy(false);}}
 }
 async function check(): Promise<void> {
  if (!operation || locked.current) {return;} locked.current = true; setBusy(true); const signal = active.current?.signal;
  try {const value = await writeOperationRequest(owner, draft, `/${operation.id}/reconcile`, {}, signal); if (!signal?.aborted && value.id === operation.id) {setOperation(value); setError(''); setPollEpoch(value => value + 1);}}
  catch {if (!signal?.aborted) {setError('Status remains unresolved. No additional GPS mutation was requested.');}}
  finally {locked.current = false; if (!signal?.aborted) {setBusy(false);}}
 }
 async function retry(): Promise<void> {
  if (!operation || disabled || locked.current) {return;} locked.current = true; setBusy(true); const signal = active.current?.signal;
  try {const value = await writeOperationRequest(owner, draft, `/${operation.id}/retry`, {generation: operation.generation}, signal); if (!signal?.aborted && value.id === operation.id) {setOperation(value); setError('');}}
  catch {if (!signal?.aborted) {setError('Retry not acknowledged. Check status before any further action; the original attempt budget remains.');}}
  finally {locked.current = false; if (!signal?.aborted) {setBusy(false);}}
 }
 async function targetAction(assetId: string, generation: number, action: 'retry' | 'reconcile', repeat = false): Promise<void> {
  const hasManualConflict = manualPendingIDs.includes(assetId) && operation?.targets?.some(target => target.assetId === assetId && target.fields.some(field => field.field === 'gps'));
  if (!operation || locked.current || (action === 'retry' && (hasManualConflict || (!repeat && disabled)))) {return;}
  const identity = repeat ? pendingTargetRetry : {operationId: operation.id, assetId, generation};
  if (identity?.operationId !== operation.id) {return;}
  confirmationVersion.current++; setPollEpoch(epoch => epoch + 1);
  locked.current = true; setBusy(true); const signal = active.current?.signal;
  try {
   if (action === 'retry') {sessionStorage.setItem(targetStorageKey, JSON.stringify(identity)); setTargetRetry(identity);}
   const value = await writeOperationRequest(owner, draft, `/${identity.operationId}/targets/${identity.assetId}/${action}`, {generation: identity.generation}, signal);
   if (!signal?.aborted && value.id === operation.id) {
    setOperation(value); setError(''); setPollEpoch(epoch => epoch + 1);
    if (action === 'retry') {setTargetRetry(null); sessionStorage.removeItem(targetStorageKey);}
   }
  } catch {if (!signal?.aborted) {setError(action === 'retry' ? `Retry acknowledgement unavailable for ${assetId}. Reconcile status or repeat the same target generation.` : `Status remains unresolved for ${assetId}. No additional mutation was requested.`);}}
  finally {locked.current = false; if (!signal?.aborted) {setBusy(false);}}
 }
 async function metadataAction(action: 'retry' | 'reconcile', repeat = false): Promise<void> {
  if (!operation?.mirror || locked.current || (action === 'retry' && !repeat && metadataDisabled)) {return;}
  const identity = repeat ? pendingMirrorRetry : {operationId: operation.id, assetId: operation.mirror.assetId, generation: operation.mirror.generation};
  if (identity?.operationId !== operation.id) {return;}
  confirmationVersion.current++; setPollEpoch(epoch => epoch + 1); locked.current = true; setBusy(true); const signal = active.current?.signal;
  try {
   if (action === 'retry') {sessionStorage.setItem(mirrorStorageKey, JSON.stringify(identity)); setMirrorRetry(identity);}
   const value = await mirrorOperationRequest(owner, draft, identity, action, signal);
   if (!signal?.aborted) {setOperation(value); setError(''); setPollEpoch(epoch => epoch + 1); if (action === 'retry') {setMirrorRetry(null); sessionStorage.removeItem(mirrorStorageKey);}}
  } catch {if (!signal?.aborted) {setError(action === 'retry' ? 'Metadata retry acknowledgement unavailable. Check status or repeat the same metadata generation.' : 'Metadata status remains unresolved. No additional mutation was requested.');}}
  finally {locked.current = false; if (!signal?.aborted) {setBusy(false);}}
 }
 return <section aria-label={hasDescription ? 'Selected fields write operation' : 'GPS write operation'} className={'space-y-2 border-t pt-3'}>
  <h4>{hasDescription ? 'Confirmed selected fields operation' : 'Confirmed GPS operation'}</h4>
  <p>{preview?.plan.version === 'mirror-preview-v4' ? 'Confirming authorizes the exact standard-field matrix and analyzed-photo metadata export shown above. Standard fields run before metadata. Manual pending choices are preserved.' : preview?.plan.version === 'stack-preview-v3' ? 'Confirming authorizes only the exact target and field matrix shown above. Manual pending choices are preserved.' : hasDescription ? 'Confirming authorizes the exact photo and selected fields shown above. Manual pending choices are preserved.' : 'Confirming authorizes only the exact photo and GPS pair shown above. Other metadata and manual pending choices are preserved.'}</p>
  {preview && !pending && operation?.plan.id !== preview.plan.id && <button type={'button'} disabled={disabled || isBusy} onClick={() => void confirm()}>{hasDescription ? 'Confirm selected fields write' : 'Confirm GPS write'}</button>}
  {pending && <button type={'button'} disabled={isBusy} onClick={() => void reconcile()}>{'Reconcile confirmation'}</button>}
  {pending && preview?.plan.id === pending.previewId && <button type={'button'} disabled={disabled || isBusy} onClick={() => void repeatConfirmation()}>{'Retry same confirmation'}</button>}
  {error && <p role={'alert'}>{error}</p>}
  {history.items.length > 0 && <label>{hasDescription ? 'Saved selected fields operations' : 'Saved GPS operations'}<select aria-label={hasDescription ? 'Saved selected fields operations' : 'Saved GPS operations'} value={operation?.id ?? ''} disabled={isBusy} onChange={event => void inspect(event.target.value)}><option value={''} disabled>{'Select operation'}</option>{history.items.map(item => <option key={item.id} value={item.id}>{`Revision ${item.draftRevision} · ${item.status} · ${item.approvedAt}`}</option>)}</select></label>}
  {history.nextCursor && <button type={'button'} disabled={isBusy} onClick={() => void historyPage(history.nextCursor)}>{hasDescription ? 'More saved selected fields operations' : 'More saved GPS operations'}</button>}
  {historyCursor && <button type={'button'} disabled={isBusy} onClick={() => void historyPage('')}>{hasDescription ? 'First saved selected fields operations' : 'First saved GPS operations'}</button>}
  {operation && <div>
   <p role={'status'}>{`${operation.plan.version === 'mirror-preview-v4' ? 'Combined' : operation.plan.version === 'stack-preview-v3' ? 'Stack' : operation.plan.version === 'standard-preview-v2' ? 'Selected fields' : 'GPS'} operation: ${operation.status}`}</p>
   <p>{`Approved revision ${operation.plan.draftRevision} · Photo ${operation.plan.targetId}`}</p>
   {operation.plan.draftRevision !== draft.revision && <p>{`This operation belongs to revision ${operation.plan.draftRevision}; current draft revision ${draft.revision} is not marked saved.`}</p>}
   <p>{`Approved at ${operation.approvedAt}${(operation.plan.version === 'stack-preview-v3' || operation.plan.version === 'mirror-preview-v4') ? '' : ` · Attempts ${operation.attempts} of 2`}`}</p>
   {(operation.plan.version === 'stack-preview-v3' || operation.plan.version === 'mirror-preview-v4') ? <StackWriteOutcome manualPendingIDs={manualPendingIDs} operation={operation} disabled={disabled || isBusy || operation.plan.draftRevision !== draft.revision || Date.parse(operation.plan.expiresAt) <= Date.now()} isBusy={isBusy} pendingRetry={pendingTargetRetry?.operationId === operation.id ? pendingTargetRetry : null} onAction={targetAction} /> : operation.plan.version === 'standard-preview-v2' ? <><StandardWriteComparison plan={operation.plan} /><StandardWriteOutcome operation={operation} /></> : operation.plan.version === 'gps-preview-v1' ? <>
   <p>{`Approved before: ${operation.plan.before.latitude ?? 'absent'}, ${operation.plan.before.longitude ?? 'absent'}`}</p>
   <p>{`Approved intended: ${operation.plan.intended.latitude}, ${operation.plan.intended.longitude}`}</p>
   <p>{`Observed GPS: ${operation.observed ? `${operation.observed.latitude ?? 'absent'}, ${operation.observed.longitude ?? 'absent'}` : 'not verified'}`}</p>
   <p>{operation.verified && operation.refreshed ? (operation.noop ? 'GPS verified unchanged; no write sent.' : 'GPS verified and local catalog updated. Verification does not prove which client changed it.') : operation.verified ? 'GPS verified upstream; local refresh pending.' : 'GPS is not verified saved.'}</p>
                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             </> : null}
   <MirrorOutcome operation={operation} isBusy={isBusy} onAction={metadataAction} pendingRetry={pendingMirrorRetry?.operationId === operation.id ? pendingMirrorRetry : null} disabled={metadataDisabled || isBusy || operation.plan.draftRevision !== draft.revision || Date.parse(operation.plan.expiresAt) <= Date.now()} />
   <p>{`Outcome: ${operation.code || ((operation.plan.version === 'stack-preview-v3' || operation.plan.version === 'mirror-preview-v4') ? operation.status : 'awaiting execution')}`}</p>
   {!operation.settled && operation.attempts > 0 && <p>{operation.plan.version === 'standard-preview-v2' ? 'Prior request completion is unknown. Another selected fields write and revision changes remain blocked.' : 'Prior request completion is unknown. Another GPS write and revision changes remain blocked.'}</p>}
   {operation.plan.version !== 'stack-preview-v3' && operation.plan.version !== 'mirror-preview-v4' && <button type={'button'} disabled={isBusy} onClick={() => void check()}>{'Check status'}</button>}
   {operation.plan.version !== 'stack-preview-v3' && operation.plan.version !== 'mirror-preview-v4' && operation.status === 'retryable' && <button type={'button'} disabled={disabled || isBusy || operation.plan.draftRevision !== draft.revision || Date.parse(operation.plan.expiresAt) <= Date.now()} onClick={() => void retry()}>{operation.plan.version === 'standard-preview-v2' ? 'Retry selected fields write' : 'Retry GPS write'}</button>}
   <details><summary>{'Private operation history'}</summary>{operation.events.map((item, index) => <p key={index}>{`${item.at} · ${item.code} · attempt ${item.attempt}`}</p>)}</details>
                </div>}
        </section>;
}

function hasActiveTargets(operation: TWriteOperation): boolean {
 if (operation.mirror && ['queued', 'writing', 'verifying'].includes(operation.mirror.status)) {return true;}
 return operation.targets ? operation.targets.some(target => ['queued', 'writing', 'verifying'].includes(target.status)) : ['queued', 'writing', 'verifying'].includes(operation.status);
}
