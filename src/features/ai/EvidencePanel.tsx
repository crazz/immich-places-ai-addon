import type {TReview, TReviewCandidate} from './reviewTypes';
import type {ReactElement} from 'react';

const sourceKinds: ReadonlyMap<string, string> = new Map([['capture_time', 'Capture time'], ['selected_album', 'Selected album label'], ['user_hint', 'User hint'], ['nearby_locations', 'Nearby location']]);
export function EvidencePanel({review, candidate}: {review: TReview; candidate: TReviewCandidate | null}): ReactElement {
 const observations = candidate ? candidate.observations.map(id => review.observations.find(item => item.id === id)) : review.observations;
 const candidateSources = candidate ? candidate.sources.flatMap(id => {
  const matches = review.sources.filter(item => item.id === id);
  return matches.length ? matches : [undefined];
 }) : review.sources;
 const sources = review.mode === 'research' ? [...candidateSources.filter(source => source?.kind !== 'answer_reference'), ...review.sources.filter(source => source.kind === 'answer_reference')] : candidateSources;
 return <details open={review.mode === 'research'} className={'space-y-2 rounded border border-(--color-border) p-3'}>
  <summary className={'cursor-pointer font-semibold'}>{'Evidence and limitations'}</summary>
  {candidate && <p className={'whitespace-pre-wrap'}>{candidate.support}</p>}
  <h4 className={'font-semibold'}>{'Observations'}</h4>
  {observations.map((item, index) => item ? <div key={item.id} className={'space-y-1'}>
   <p className={'text-xs'}>{item.kind === 'visual' ? 'Model / visual evidence; no independent verification.' : 'Model interpretation of supplied context; not external verification.'}</p>
   <p className={'whitespace-pre-wrap'}>{item.text}</p>
                                            </div> : <p key={index}>{'Referenced observation unavailable.'}</p>)}
  <h4 className={'font-semibold'}>{review.mode === 'research' ? 'References' : 'Stored source provenance'}</h4>
  {!sources.length && <p>{review.mode === 'research' ? 'No reference links were included.' : 'No external context sources were supplied.'}</p>}
  {sources.map((source, index) => source ? <div key={`${source.kind}:${source.id}`} className={'space-y-1'}>
   {source.kind === 'answer_reference' ? <>
    {source.url ? <a href={source.url} target={'_blank'} rel={'noopener noreferrer'} className={'break-words underline'}>{source.title || 'View reference'}</a> : <p>{source.title || 'Reference link unavailable'}</p>}
    <p className={'whitespace-pre-wrap break-words'}>{source.relevance}</p>
    <p className={'text-xs'}>{'AI-provided reference; not independently verified.'}</p>
                                         </> : <>
   <p>{`${sourceKinds.get(source.kind) || 'Context source'} · ${source.id}`}</p>
   <p>{source.lineage === 'unknown' ? 'Lineage: unknown; not independent corroboration.' : `Recorded lineage: ${source.lineage}; no verification claim is implied.`}</p>
   <p className={'text-xs'}>{'Only retained provenance is shown. Source content and public citations are not reconstructed.'}</p>
                                               </>}
                                           </div> : <p key={index}>{'Referenced source unavailable.'}</p>)}
  {review.omissions.map((text, index) => <p key={index}>{`Context limitation: ${text}`}</p>)}
  {review.warnings.length > 0 && <><h4 className={'font-semibold'}>{'Warnings'}</h4><ul className={'list-disc pl-5'}>{review.warnings.map((text, index) => <li key={index}>{text}</li>)}</ul></>}
        </details>;
}
