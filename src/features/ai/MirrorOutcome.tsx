import {MirrorComparison} from './MirrorComparison';

import type {TWriteOperation} from './writeOperationTypes';
import type {ReactElement} from 'react';

export function MirrorOutcome({operation, disabled, isBusy, onAction, pendingRetry}: {operation: TWriteOperation; disabled?: boolean; isBusy?: boolean; onAction?: (action: 'retry' | 'reconcile', repeat?: boolean) => Promise<void>; pendingRetry?: {generation: number} | null}): ReactElement | null {
 if (operation.plan.version !== 'mirror-preview-v4' || !operation.mirror) {return null;}
 const mirror = operation.mirror;
 return <section aria-label={'Metadata step outcome'} className={'space-y-2 rounded border p-2'}>
  <MirrorComparison mirror={operation.plan.mirror} targetId={mirror.assetId} />
  <p>{`Metadata step: ${mirror.status}`}</p><p>{`Metadata outcome: ${mirror.code || 'awaiting execution'}`}</p>
  <p>{`Metadata attempts ${mirror.attempts} of 2 · Generation ${mirror.generation}`}</p>
  <p>{'Standard-field outcomes remain independent. Local direction and translations are preserved.'}</p>
  {mirror.verified && <p>{mirror.status === 'conflict' ? 'Approved metadata was previously verified; the latest observation conflicts with it.' : 'Approved metadata observed upstream; this does not prove native display or file updates.'}</p>}
  {!mirror.settled && <p>{'Metadata sender completion is unresolved; another metadata write is not permitted.'}</p>}
  {['writing', 'verifying'].includes(mirror.status) && mirror.generation >= 1 && <button type={'button'} disabled={isBusy} onClick={() => void onAction?.('reconcile')}>{'Check metadata status'}</button>}
  {mirror.status === 'retryable' && !pendingRetry && !mirror.verified && mirror.settled && mirror.attempts < 2 && mirror.generation >= 1 && <button type={'button'} disabled={disabled || isBusy} onClick={() => void onAction?.('retry')}>{'Retry metadata only'}</button>}
  {pendingRetry && <button type={'button'} disabled={isBusy} onClick={() => void onAction?.('retry', true)}>{`Repeat metadata retry generation ${pendingRetry.generation}`}</button>}
  <details><summary>{'Private metadata history'}</summary><pre aria-label={'Observed metadata'} tabIndex={0} className={'whitespace-pre-wrap break-words'}>{JSON.stringify(mirror.observed, null, 2)}</pre>{mirror.events.map((event, index) => <p key={index}>{`${event.at} · ${event.code} · attempt ${event.attempt} · generation ${event.generation}`}</p>)}</details>
        </section>;
}
