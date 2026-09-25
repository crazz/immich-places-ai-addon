import {useCallback, useEffect, useRef, useState} from 'react';

import {AIRequestError} from './aiRequest';
import {fetchWriteHistory, writeOperationRequest} from './writeOperationApi';
import {isWriteConfirmation} from './writeOperationTypes';

import type {TDraft} from './draftTypes';
import type {TWriteConfirmation, TWriteHistory, TWriteOperation} from './writeOperationTypes';
import type {TWritePreview} from './writePreviewTypes';
import type {ReactElement} from 'react';

type TProps = {owner: string; draft: TDraft; preview: TWritePreview | null; disabled?: boolean; onVerified?: () => void};

export function WriteOperation(props: TProps): ReactElement {
 return <OperationSession key={`${props.owner}:${props.draft.id}`} {...props} />;
}

function OperationSession({owner, draft, preview, disabled, onVerified}: TProps): ReactElement {
 const [operation, setOperation] = useState<TWriteOperation | null>(null);
 const [pending, setPending] = useState<TWriteConfirmation | null>(null);
 const [history, setHistory] = useState<TWriteHistory>({items: [], nextCursor: ''});
 const [historyCursor, setHistoryCursor] = useState('');
 const [error, setError] = useState(''); const [isBusy, setBusy] = useState(false);
 const [pollEpoch, setPollEpoch] = useState(0);
 const operationID = operation?.id; const operationStatus = operation?.status;
 const active = useRef<AbortController | null>(null); const locked = useRef(false);
 const confirmationVersion = useRef(0);
 const storageKey = `ai-write-confirmation:${owner}:${draft.id}`;
 const notified = useRef('');
 useEffect(() => {
  if (operation?.verified && operation.refreshed && notified.current !== operation.id) {notified.current = operation.id; onVerified?.();}
 }, [operation, onVerified]);
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
  if (!operationID || !operationStatus || !['queued', 'writing', 'verifying'].includes(operationStatus)) {return;}
  const id = operationID; const controller = new AbortController(); let reads = 0; let timer: ReturnType<typeof setTimeout>;
  const poll = async (): Promise<void> => {
   try {
    const value = await writeOperationRequest(owner, draft, `/${id}`, undefined, controller.signal);
    if (controller.signal.aborted || value.id !== id) {return;} setOperation(value); reads++;
    if (['queued', 'writing', 'verifying'].includes(value.status) && reads < 30) {timer = setTimeout(() => void poll(), 2_000);}
   } catch {if (!controller.signal.aborted) {setError('Status read unavailable. Use Check status; no additional GPS mutation was requested.');}}
  };
  timer = setTimeout(() => void poll(), 2_000);
  return () => {controller.abort(); clearTimeout(timer);};
 }, [owner, draft, operationID, operationStatus, pollEpoch]);
 function handleConfirmationRejection(failure: unknown): boolean {
  if (!(failure instanceof AIRequestError) || !['WRITE_DISABLED', 'PREVIEW_EXPIRED', 'PREVIEW_CONSUMED', 'DRAFT_CONFLICT', 'PLAN_CONFLICT', 'TARGET_BUSY', 'INVALID_WRITE', 'IDEMPOTENCY_CONFLICT'].includes(failure.code)) {return false;}
  setPending(null); try {sessionStorage.removeItem(storageKey);} catch { /* A definitive rejection grants no authority. */ }
  setError(failure.code === 'WRITE_DISABLED' ? 'GPS writing is disabled. Saved review and status remain available.' : 'Confirmation was rejected. Inspect saved GPS history and create a fresh comparison after resolving the conflict.');
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
  const signal = active.current?.signal; setBusy(true);
  try {const value = await writeOperationRequest(owner, draft, `/${id}`, undefined, signal); if (!signal?.aborted && value.id === id) {setOperation(value);}}
  catch {if (!signal?.aborted) {setError('Saved GPS operation unavailable.');}}
  finally {if (!signal?.aborted) {setBusy(false);}}
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
 return <section aria-label={'GPS write operation'} className={'space-y-2 border-t pt-3'}>
  <h4>{'Confirmed GPS operation'}</h4>
  <p>{'Confirming authorizes only the exact photo and GPS pair shown above. Other metadata and manual pending choices are preserved.'}</p>
  {preview && !pending && operation?.plan.id !== preview.plan.id && <button type={'button'} disabled={disabled || isBusy} onClick={() => void confirm()}>{'Confirm GPS write'}</button>}
  {pending && <button type={'button'} disabled={isBusy} onClick={() => void reconcile()}>{'Reconcile confirmation'}</button>}
  {pending && preview?.plan.id === pending.previewId && <button type={'button'} disabled={disabled || isBusy} onClick={() => void repeatConfirmation()}>{'Retry same confirmation'}</button>}
  {error && <p role={'alert'}>{error}</p>}
  {history.items.length > 0 && <label>{'Saved GPS operations'}<select aria-label={'Saved GPS operations'} value={operation?.id ?? ''} disabled={isBusy} onChange={event => void inspect(event.target.value)}><option value={''} disabled>{'Select operation'}</option>{history.items.map(item => <option key={item.id} value={item.id}>{`Revision ${item.draftRevision} · ${item.status} · ${item.approvedAt}`}</option>)}</select></label>}
  {history.nextCursor && <button type={'button'} disabled={isBusy} onClick={() => void historyPage(history.nextCursor)}>{'More saved GPS operations'}</button>}
  {historyCursor && <button type={'button'} disabled={isBusy} onClick={() => void historyPage('')}>{'First saved GPS operations'}</button>}
  {operation && <div>
   <p role={'status'}>{`GPS operation: ${operation.status}`}</p>
   <p>{`Approved revision ${operation.plan.draftRevision} · Photo ${operation.plan.targetId}`}</p>
   <p>{`Approved at ${operation.approvedAt} · Attempts ${operation.attempts} of 2`}</p>
   <p>{`Approved before: ${operation.plan.before.latitude ?? 'absent'}, ${operation.plan.before.longitude ?? 'absent'}`}</p>
   <p>{`Approved intended: ${operation.plan.intended.latitude}, ${operation.plan.intended.longitude}`}</p>
   <p>{`Observed GPS: ${operation.observed ? `${operation.observed.latitude ?? 'absent'}, ${operation.observed.longitude ?? 'absent'}` : 'not verified'}`}</p>
   <p>{operation.verified && operation.refreshed ? (operation.noop ? 'GPS verified unchanged; no write sent.' : 'GPS verified and local catalog updated. Verification does not prove which client changed it.') : operation.verified ? 'GPS verified upstream; local refresh pending.' : 'GPS is not verified saved.'}</p>
   <p>{`Outcome: ${operation.code || 'awaiting execution'}`}</p>
   {!operation.settled && operation.attempts > 0 && <p>{'Prior request completion is unknown. Another GPS write and revision changes remain blocked.'}</p>}
   <button type={'button'} disabled={isBusy} onClick={() => void check()}>{'Check status'}</button>
   {operation.status === 'retryable' && <button type={'button'} disabled={disabled || isBusy || operation.plan.draftRevision !== draft.revision || Date.parse(operation.plan.expiresAt) <= Date.now()} onClick={() => void retry()}>{'Retry GPS write'}</button>}
   <details><summary>{'Private operation history'}</summary>{operation.events.map((item, index) => <p key={index}>{`${item.at} · ${item.code} · attempt ${item.attempt}`}</p>)}</details>
                </div>}
        </section>;
}
