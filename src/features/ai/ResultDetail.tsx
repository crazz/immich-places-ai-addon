import {useCallback, useMemo} from 'react';

import {ProposalReview} from './ProposalReview';
import {ResultImage} from './ResultImage';
import {fetchResult} from './resultsApi';
import {parseReview} from './reviewParser';
import {useResultRead} from './useResultRead';

import type {TResultReference} from './resultDetailTypes';
import type {ReactElement} from 'react';

export function ResultDetail({owner, reference}: {owner: string; reference: TResultReference}): ReactElement {
 const load = useCallback(async (signal: AbortSignal) => fetchResult(reference, signal), [reference]);
 const {data, loading: isLoading, error, refresh} = useResultRead(owner, JSON.stringify(reference), load);
 const review = useMemo(() => data ? parseReview(data) : null, [data]);
 return <section aria-label={'Saved result detail'} className={'space-y-3 break-words'}>
  {isLoading && <p role={'status'}>{'Loading result…'}</p>}
  {error && <><p role={'alert'}>{error}</p><button type={'button'} onClick={refresh}>{'Retry result'}</button></>}
  {data && <>
   <ResultImage owner={owner} entry={data.entry} />
   <h2 className={'font-semibold'}>{data.entry.label || 'Saved AI run'}</h2>
   <p>{`Execution: ${data.entry.executionState}`}</p>
   <p>{`Proposal: ${data.entry.proposalOutcome || 'none'}`}</p>
   <p>{'Review: unreviewed · Write: not requested'}</p>
   {data.entry.proposalOutcome === 'unknown' && <p>{'No location was established.'}</p>}
   {!data.proposal && <p>{'This run has no saved proposal.'}</p>}
   {data.proposal && (review ? <ProposalReview key={`${owner}:${data.entry.id}`} review={review} /> : <p role={'alert'}>{'Stored proposal unavailable: invalid geometry, evidence or language contract.'}</p>)}
   <details><summary className={'cursor-pointer'}>{'Original run provenance'}</summary>
    <dl className={'grid grid-cols-[auto_minmax(0,1fr)] gap-2 pt-2 text-xs'}>
     {Object.entries({Mode: data.entry.mode, Model: data.provenance.Model, Revision: data.provenance.Revision, Languages: data.provenance.Languages.join(', '), Primary: data.provenance.PrimaryLanguage, Prompt: data.provenance.PromptVersion || 'Unavailable', Schema: data.provenance.SchemaVersion || 'Unavailable', Validation: data.provenance.ValidationVersion || 'Unavailable', Job: data.entry.jobId, Asset: data.entry.assetId, Analysis: data.entry.analysisId || 'None'}).map(([label, value]) => <div key={label} className={'contents'}><dt className={'font-semibold'}>{label}</dt><dd className={'break-all'}>{value}</dd></div>)}
    </dl>
   </details>
           </>}
        </section>;
}
