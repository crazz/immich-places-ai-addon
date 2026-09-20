'use client';
import {Description} from '@radix-ui/react-dialog';
import {useEffect, useRef, useState} from 'react';

import {DialogShell} from '@/shared/components/DialogShell';

import {BatchLaunch} from './BatchLaunch';
import {JobList} from './JobList';
import {JobProgressPanel} from './JobProgressPanel';
import {isUUID} from './selectionTypes';

import type {TRerunChoice} from './jobTypes';
import type {TBatchScope} from './selectionIntent';
import type {ReactElement} from 'react';

type TProps = {owner: string; input: TBatchScope};

function WorkspaceSession({owner, input}: TProps): ReactElement {
 const [isOpen, setIsOpen] = useState(false);
 const entry = useRef<HTMLButtonElement>(null);
 const wasOpen = useRef(false);
 useEffect(() => {
  if (wasOpen.current && !isOpen) {entry.current?.focus();}
  wasOpen.current = isOpen;
 }, [isOpen]);
 const [view, setView] = useState<'launch' | 'jobs' | 'progress'>('launch');
 const [jobID, setJobID] = useState<string | null>(null);
 const [rerun, setRerun] = useState<{choice: TRerunChoice; assetIDs: string[]} | null>(null);
 useEffect(() => {
  const restore = (): void => {
   const id = new URL(window.location.href).searchParams.get('aiJob');
   setJobID(isUUID(id) ? id : null);
   setIsOpen(isUUID(id));
   setView(isUUID(id) ? 'progress' : 'launch');
   setRerun(null);
  };
  restore(); window.addEventListener('popstate', restore);
  return () => window.removeEventListener('popstate', restore);
 }, []);
 const navigate = (id: string | null): void => {
  const url = new URL(window.location.href);
  if (id) {url.searchParams.set('aiJob', id);} else {url.searchParams.delete('aiJob');}
  window.history.pushState(null, '', url);
  setJobID(id);
 };
 const openJob = (id: string): void => {navigate(id); setRerun(null); setView('progress'); setIsOpen(true);};
 const launchInput: TBatchScope = rerun ? {scope: {view: 'all', gpsFilter: 'all', hiddenFilter: 'visible'}, selected: rerun.assetIDs, page: [], blockedReason: ''} : input;
 return <>
  <button ref={entry} type={'button'} className={'rounded border border-(--color-border) px-3 py-1 text-sm'} onClick={() => setIsOpen(true)}>{'AI Locate'}</button>
  <DialogShell isOpen={isOpen} onClose={() => {setIsOpen(false); navigate(null); setRerun(null); setView('launch');}} title={'AI Locate'} maxWidth={'review'} subtitle={'Analyze images and review durable jobs. Closing this window does not stop a job.'}>
   {isOpen && <div className={'max-h-[min(75dvh,calc(100dvh-7rem))] min-w-0 space-y-5 overflow-y-auto p-4 text-sm wrap-anywhere sm:p-6 [&_button]:cursor-pointer [&_button]:rounded [&_button]:border [&_button]:border-(--color-border) [&_button]:px-3 [&_button]:py-2 [&_button:disabled]:cursor-default [&_button:disabled]:opacity-50 [&_input:not([type=checkbox])]:w-full [&_input:not([type=checkbox])]:min-w-0 [&_input:not([type=checkbox])]:rounded [&_input:not([type=checkbox])]:border [&_input:not([type=checkbox])]:border-(--color-border) [&_input:not([type=checkbox])]:p-2 [&_select]:w-full [&_select]:min-w-0 [&_select]:rounded [&_select]:border [&_select]:border-(--color-border) [&_select]:p-2 [&_textarea]:w-full [&_textarea]:min-w-0 [&_textarea]:rounded [&_textarea]:border [&_textarea]:border-(--color-border) [&_textarea]:p-2'}>
    <Description className={'sr-only'}>{'Preview and authorize image analysis, observe saved jobs, or cancel unfinished work. Closing this window does not stop a job.'}</Description>
    <nav aria-label={'AI workspace'} className={'flex gap-2'}>
     <button type={'button'} onClick={() => {navigate(null); setRerun(null); setView('launch');}}>{'New analysis'}</button>
     <button type={'button'} onClick={() => {navigate(null); setRerun(null); setView('jobs');}}>{'Saved jobs'}</button>
    </nav>
    {view === 'launch' && <>
     {rerun && <p>{`${rerun.choice.kind === 'retry-failed' ? 'Retry failed assets' : 'Reanalyze selected assets'} · Parent job ${rerun.choice.parentJobId}`}</p>}
     <BatchLaunch input={launchInput} rerun={rerun?.choice} onSubmittedAction={job => openJob(job.id)} />
                          </>}
    {view === 'jobs' && <JobList owner={owner} onOpenAction={openJob} />}
    {view === 'progress' && jobID && <JobProgressPanel owner={owner} id={jobID} onRerunAction={(choice, assetIDs) => {navigate(null); setRerun({choice, assetIDs}); setView('launch');}} />}
              </div>}
  </DialogShell>
        </>;
}

export function AIWorkspace(props: TProps): ReactElement {return <WorkspaceSession key={props.owner} {...props} />;}
