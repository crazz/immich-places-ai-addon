import type {TWriteOperation} from './writeOperationTypes';
import type {ReactElement} from 'react';

export function StandardWriteOutcome({operation}: {operation: TWriteOperation}): ReactElement {
 return <div>
  {operation.fields?.map(item => <div key={item.field}>
   <p>{`${item.field === 'gps' ? 'GPS' : 'Description'}: ${item.status}`}</p>
   {item.wasVerified && item.status !== 'verified' && <p>{`${item.field === 'gps' ? 'GPS' : 'Description'} was previously verified; this is retained evidence, not current verification.`}</p>}
   {item.gps && <p>{`Observed GPS: ${item.gps.latitude ?? 'absent'}, ${item.gps.longitude ?? 'absent'}`}</p>}
   {item.description && <><p>{`Observed description presence: ${item.description.presence}`}</p><pre aria-label={'Observed description'} tabIndex={0} className={'whitespace-pre-wrap break-words'}>{item.description.value}</pre></>}
                                 </div>)}
  {!operation.verified && <p>{'Selected fields are not fully verified saved.'}</p>}
  {!operation.verified && operation.refreshed && operation.fields?.some(item => item.field === 'gps' && item.status === 'verified') && <p>{'GPS verified and local catalog updated; selected fields remain incomplete.'}</p>}
  {operation.verified && operation.plan.fields.includes('gps') && <p>{operation.refreshed ? 'Selected fields verified and GPS catalog updated.' : 'Selected fields verified upstream; GPS catalog refresh pending.'}</p>}
  {operation.verified && !operation.plan.fields.includes('gps') && <p>{'Description verified; no GPS catalog refresh is needed.'}</p>}
        </div>;
}
