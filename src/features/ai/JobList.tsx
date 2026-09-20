import {useEffect, useState} from 'react';

import {fetchJobs} from './jobApi';

import type {TJobPage} from './jobTypes';
import type {ReactElement} from 'react';

type TProps = {owner: string; onOpenAction: (id: string) => void};

function JobListSession({onOpenAction}: TProps): ReactElement {
 const [cursor, setCursor] = useState<string>();
 const [generation, setGeneration] = useState(0);
 const [page, setPage] = useState<TJobPage | null>(null);
 const [error, setError] = useState('');
 useEffect(() => {
  const controller = new AbortController(); setPage(null); setError('');
  fetchJobs(cursor, controller.signal).then(result => {
   if (!controller.signal.aborted) {setPage(result);}
  }).catch(() => {
   if (!controller.signal.aborted) {setError('Could not load jobs. Try refreshing.');}
  });
  return () => controller.abort();
 }, [cursor, generation]);
 return <section aria-label={'Saved AI jobs'} className={'space-y-3'}>
  <button type={'button'} onClick={() => {setCursor(undefined); setGeneration(value => value + 1);}}>{'Refresh jobs'}</button>
  {error && <p role={'alert'}>{error}</p>}
  {!page && !error && <p role={'status'}>{'Loading saved jobs…'}</p>}
  {page?.items.length === 0 && <p>{'No saved jobs.'}</p>}
  <ul className={'space-y-2'}>{page?.items.map(job => <li key={job.id} className={'rounded border border-(--color-border) p-2'}>
   <button type={'button'} aria-label={`Open job ${job.id}`} onClick={() => onOpenAction(job.id)}>{`${new Date(job.createdAt / 1000000).toLocaleString()} · ${job.configuration.mode} · ${job.counts.total} assets`}</button>
   <p>{job.blocked ? 'Blocked' : job.canceled ? 'Cancellation requested' : Object.entries(job.counts).filter(([key]) => key !== 'total').map(([key, count]) => `${key}: ${count}`).join(' · ')}</p>
                                                      </li>)}
  </ul>
  {page?.nextCursor && <button type={'button'} onClick={() => setCursor(page.nextCursor ?? undefined)}>{'Older jobs'}</button>}
        </section>;
}

export function JobList(props: TProps): ReactElement {return <JobListSession key={props.owner} {...props} />;}
