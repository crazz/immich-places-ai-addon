import {useState} from 'react';

import type {TTranslationRun} from './translationTypes';
import type {ReactElement} from 'react';

export function TranslationSuggestions({run, disabled, onAdopt, onRetry}: {run: TTranslationRun; disabled: boolean; onAdopt: (languages: string[]) => void; onRetry: (languages: string[]) => void}): ReactElement {
 const [selected,setSelected] = useState<string[]>([]);
 const [retry,setRetry] = useState<string[]>([]);
 return <div aria-label={'Translation suggestions'} className={'space-y-2'}>
  <p>{'Basis used for these suggestions:'}</p><p className={'whitespace-pre-wrap break-words'}>{run.request.basis}</p>
  {run.items.map(item => <div key={item.language}>
   <p role={'status'}>{`${item.language}: ${item.state}`}</p>{item.text && <p className={'whitespace-pre-wrap break-words'}>{item.text}</p>}
   <label><input type={'checkbox'} disabled={disabled || item.state !== 'complete'} checked={selected.includes(item.language)} onChange={event => setSelected(values => event.target.checked ? [...values,item.language] : values.filter(value => value !== item.language))} />{`Adopt ${item.language}`}</label>
   {['unavailable','failed','canceled','interrupted'].includes(item.state) && <label><input type={'checkbox'} disabled={disabled} checked={retry.includes(item.language)} onChange={event => setRetry(values => event.target.checked ? [...values,item.language] : values.filter(value => value !== item.language))} />{`Retry ${item.language}`}</label>}
                         </div>)}
  <button type={'button'} disabled={disabled || selected.length === 0} onClick={() => onAdopt(selected)}>{'Adopt selected translations'}</button>
  <button type={'button'} disabled={disabled || retry.length === 0 || run.items.some(item => ['queued','reserved'].includes(item.state))} onClick={() => onRetry(retry)}>{'Review retry languages'}</button>
        </div>;
}
