import {StackWriteComparison} from './StackWriteComparison';
import {StandardWriteOutcome} from './StandardWriteOutcome';

import type {TWriteOperation} from './writeOperationTypes';
import type {ReactElement} from 'react';

export function StackWriteOutcome({operation, disabled, isBusy, pendingRetry, manualPendingIDs = [], onAction}: {operation: TWriteOperation; disabled?: boolean; manualPendingIDs?: string[]; isBusy?: boolean; pendingRetry?: {assetId: string; generation: number} | null; onAction?: (assetId: string, generation: number, action: 'retry' | 'reconcile', repeat?: boolean) => Promise<void>}): ReactElement | null {
 if (operation.plan.version !== 'stack-preview-v3' && operation.plan.version !== 'mirror-preview-v4') {return null;}
 return <div className={'space-y-3'}>
  <StackWriteComparison plan={operation.plan} />
  {operation.targets?.map(target => {
   const fieldLabel = target.fields.length > 1 ? 'Selected fields' : target.fields[0]?.field === 'description' ? 'Description' : 'GPS';
   const hasManualConflict = target.fields.some(field => field.field === 'gps') && manualPendingIDs.includes(target.assetId);
   return <section key={target.assetId} aria-label={`Outcome for ${target.assetId}`} className={'rounded border p-2'}>
   <p className={'break-all'}>{`Photo ${target.assetId}`}</p>
   <p>{`Target status: ${target.status}`}</p><p>{`Outcome: ${target.code || 'awaiting execution'}`}</p>
   <p>{`Attempts ${target.attempts} of 2 · Generation ${target.generation}`}</p>
   <StandardWriteOutcome operation={{...operation, ...target}} />
   {target.verified && target.noop && <p>{`${fieldLabel} verified unchanged; no write sent for this target.`}</p>}
   {!target.settled && <p>{'Sender completion unresolved; target exclusion is retained.'}</p>}
   {(!target.verified || !target.settled || !target.refreshed) && target.generation >= 1 && <button type={'button'} disabled={isBusy} onClick={() => void onAction?.(target.assetId, target.generation, 'reconcile')}>{`Check status for ${target.assetId}`}</button>}
   {hasManualConflict && <p>{'Resolve this target’s pending manual GPS before any retry. Read-only status remains available.'}</p>}
   {target.status === 'retryable' && !pendingRetry && <button type={'button'} disabled={disabled || hasManualConflict || !target.settled || target.attempts >= 2 || target.generation < 1} onClick={() => void onAction?.(target.assetId, target.generation, 'retry')}>{`Retry ${fieldLabel === 'GPS' ? fieldLabel : fieldLabel.toLowerCase()} for ${target.assetId}`}</button>}
   {pendingRetry?.assetId === target.assetId && <button type={'button'} disabled={isBusy || hasManualConflict} onClick={() => void onAction?.(target.assetId, pendingRetry.generation, 'retry', true)}>{`Repeat retry generation ${pendingRetry.generation} for ${target.assetId}`}</button>}
   <details><summary>{'Private target history'}</summary>{target.events.map((event, index) => <p key={index}>{`${event.at} · ${event.code} · attempt ${event.attempt}`}</p>)}</details>
          </section>;
  })}
        </div>;
}
