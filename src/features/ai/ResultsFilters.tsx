import type {TResultFilters, TResultQuery} from './resultTypes';
import type {ReactElement} from 'react';

export function ResultsFilters({query, onApplyAction}: {query: TResultQuery; onApplyAction: (filters: TResultFilters) => void}): ReactElement {
 return <details><summary className={'cursor-pointer'}>{'Filter results'}</summary>
  <form
key={JSON.stringify(query)} className={'mt-3 grid grid-cols-2 gap-3'} onSubmit={event => {
   event.preventDefault();
   const filters: TResultFilters = {};
   for (const [key, value] of new FormData(event.currentTarget)) {if (typeof value === 'string' && value) {Object.assign(filters, {[key]: value});}}
   if (filters.undated === 'true') {delete filters.startDate; delete filters.endDate;}
   onApplyAction(filters);
  }}>
   <label>{'Execution'}<select name={'state'} defaultValue={query.state || ''}><option value={''}>{'All'}</option>{['succeeded', 'failed', 'canceled'].map(value => <option key={value}>{value}</option>)}</select></label>
   <label>{'Outcome'}<select name={'outcome'} defaultValue={query.outcome || ''}><option value={''}>{'All'}</option>{['located', 'ambiguous', 'unknown'].map(value => <option key={value}>{value}</option>)}</select></label>
   <label>{'Captured from'}<input name={'startDate'} type={'date'} defaultValue={query.startDate} /></label>
   <label>{'Captured through'}<input name={'endDate'} type={'date'} defaultValue={query.endDate} /></label>
   <label className={'col-span-2'}>{'Capture date'}<select name={'undated'} defaultValue={query.undated || ''}><option value={''}>{'Any'}</option><option value={'true'}>{'Unknown only (ignores date range)'}</option></select></label>
   {([['assetId', 'Asset ID'], ['jobId', 'Job ID'], ['albumId', 'Selected album ID at launch']] as const).map(([name, label]) => <label className={'col-span-2'} key={name}>{label}<input name={name} maxLength={36} defaultValue={query[name]} /></label>)}
   <p className={'col-span-2 text-xs'}>{'Dates and albums use retained launch facts. Older runs may have unknown capture dates or no retained album selection.'}</p>
   <button type={'submit'}>{'Apply filters'}</button><button type={'button'} onClick={() => onApplyAction({})}>{'Clear filters'}</button>
  </form>
        </details>;
}
