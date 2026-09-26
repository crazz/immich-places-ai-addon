import type {TStackTarget} from './writePreviewTypes';
import type {ReactElement} from 'react';

export function StandardWriteComparison({plan}: {plan: Pick<TStackTarget, 'fields' | 'description'> & Partial<Pick<TStackTarget, 'before' | 'intended'>>}): ReactElement {
 return <div className={'min-w-0 space-y-2'}>
  <p>{`Selected fields: ${plan.fields.join(', ')}`}</p>
  {plan.fields.includes('gps') && plan.before && plan.intended && ([['Before latitude', plan.before.latitude], ['Before longitude', plan.before.longitude], ['Proposed latitude', plan.intended.latitude], ['Proposed longitude', plan.intended.longitude]] as const).map(([label, value]) => <label key={label} className={'grid! gap-1 text-sm'}>{label}<input readOnly value={value ?? 'absent'} className={'w-full rounded border p-2'} /></label>)}
  {plan.description && <>
   <p>{`Description language: ${plan.description.language} · Policy: ${plan.description.policy}`}</p>
   <p>{`Before description presence: ${plan.description.before.presence}`}</p>
   <p>{'Before description'}</p><pre aria-label={'Before description'} tabIndex={0} className={'whitespace-pre-wrap break-words rounded border p-2'}>{plan.description.before.value}</pre>
   <p>{'Proposed description'}</p><pre aria-label={'Proposed description'} tabIndex={0} className={'whitespace-pre-wrap break-words rounded border p-2'}>{plan.description.intended}</pre>
                       </>}
        </div>;
}
