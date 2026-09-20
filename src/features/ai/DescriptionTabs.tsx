import {useId, useRef, useState} from 'react';

import type {TReview} from './reviewTypes';
import type {ReactElement} from 'react';

export function DescriptionTabs({review, focus}: {review: TReview; focus: string | null}): ReactElement {
 const [language, setLanguage] = useState(review.primaryLanguage);
 const id = useId(); const tabs = useRef(new Map<string, HTMLButtonElement>());
 const active = review.languages.includes(language) ? language : review.primaryLanguage;
 const description = review.descriptions.find(item => item.language === active);
 return <section aria-label={'Requested descriptions'} className={'space-y-3'}>
  <h3 className={'font-semibold'}>{'Descriptions'}</h3>
  <div role={'tablist'} aria-label={'Description languages'} className={'flex flex-wrap gap-2'}>
   {review.languages.map((tag, index) => <button
key={tag} ref={element => {if (element) {tabs.current.set(tag, element);} else {tabs.current.delete(tag);}}} type={'button'} role={'tab'} id={`${id}-${tag}`} aria-controls={`${id}-panel`} aria-selected={tag === active} tabIndex={tag === active ? 0 : -1} onClick={() => setLanguage(tag)} onKeyDown={event => {
    const next = event.key === 'ArrowRight' ? (index + 1) % review.languages.length : event.key === 'ArrowLeft' ? (index + review.languages.length - 1) % review.languages.length : event.key === 'Home' ? 0 : event.key === 'End' ? review.languages.length - 1 : -1;
    if (next < 0) {return;}
    event.preventDefault(); const selected = review.languages[next]; setLanguage(selected); tabs.current.get(selected)?.focus();
   }}>{`${tag}${tag === review.primaryLanguage ? ' · Primary' : ''}`}
                                         </button>)}
  </div>
  <div role={'tabpanel'} id={`${id}-panel`} aria-labelledby={`${id}-${active}`} tabIndex={0} className={'space-y-2 rounded border border-(--color-border) p-3'}>
   {description ? <>
    <p>{`Status: ${description.status}`}</p>
    <p>{description.basis === 'scene_only' ? 'Basis: visible scene only' : `Basis: candidate ${description.candidateId}`}</p>
    {description.candidateId && focus && description.candidateId !== focus && <p className={'text-xs'}>{`Description refers to ${description.candidateId}, not the currently inspected alternative.`}</p>}
    {description.text !== null && <p lang={description.language} dir={'auto'} className={'whitespace-pre-wrap'}>{description.text}</p>}
    {description.reason !== null && <p>{description.reason}</p>}
                  </> : <p role={'alert'}>{'Description unavailable: invalid stored language contract.'}</p>}
  </div>
        </section>;
}
