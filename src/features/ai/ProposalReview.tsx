import {useId, useState} from 'react';

import {CandidateFacts} from './CandidateFacts';
import {DescriptionTabs} from './DescriptionTabs';
import {EvidencePanel} from './EvidencePanel';
import {ReviewMapDynamic} from './ReviewMapDynamic';

import type {TReview} from './reviewTypes';
import type {ReactElement} from 'react';

export function ProposalReview({review}: {review: TReview}): ReactElement {
 const [focus, setFocus] = useState(review.initialCandidate); const mapId = useId();
 const candidate = review.candidates.find(item => item.id === focus) || null;
 return <div className={'grid items-start gap-5 lg:grid-cols-2'}>
  <div id={mapId} className={'lg:sticky lg:top-0'}><ReviewMapDynamic review={review} focus={focus} /></div>
  <div className={'min-w-0 space-y-4'}>
   <p role={'status'}>{focus ? `Inspecting ${focus}. No approval has been recorded.` : 'Viewing all alternatives. No candidate is approved.'}</p>
   {!review.candidates.length && <p>{'No camera position or candidate was proposed.'}</p>}
   {review.candidates.map(item => <article key={item.id} className={'space-y-2 rounded border border-(--color-border) p-3'}>
    <h3 className={'font-semibold'}>{item.name}</h3>
    {(item.locality || item.country) && <p>{[item.locality, item.country].filter(Boolean).join(' · ')}</p>}
    <button type={'button'} aria-pressed={focus === item.id} aria-controls={mapId} onClick={() => setFocus(item.id)}>{`Inspect candidate ${item.id}: ${item.name}`}</button>
    <CandidateFacts candidate={item} isResearch={review.mode === 'research'} />
                                  </article>)}
   {focus && <button type={'button'} onClick={() => setFocus(null)}>{'Show all alternatives'}</button>}
   <DescriptionTabs review={review} focus={focus} />
   <EvidencePanel review={review} candidate={candidate} />
  </div>
        </div>;
}
