'use client';
import {Description} from '@radix-ui/react-dialog';
import {useEffect, useRef, useState} from 'react';

import {DialogShell} from '@/shared/components/DialogShell';

import {ResultDetail} from './ResultDetail';
import {ResultList} from './ResultList';
import {resultLocation, resultURL} from './resultLocation';
import {ResultsFilters} from './ResultsFilters';

import type {TResultLocation} from './resultLocation';
import type {ReactElement} from 'react';

function ResultsSession({owner, manualPendingIDs, onVerified}: {owner: string; manualPendingIDs?: string[]; onVerified?: () => void}): ReactElement {
 const [state, setState] = useState<TResultLocation>({open: false, query: {}, reference: null});
 const [isDirty, setIsDirty] = useState(false); const [pendingNavigation, setPendingNavigation] = useState<TResultLocation | null>(null);
 const entry = useRef<HTMLButtonElement>(null);
 useEffect(() => setState(resultLocation(new URL(window.location.href))), []);
 const row = useRef(''); const wasOpen = useRef(false); const hadDetail = useRef(false);
 useEffect(() => {
  const restore = (): void => {
   const next = resultLocation(new URL(window.location.href));
   if (isDirty) {setPendingNavigation(next); window.history.pushState(null, '', resultURL(new URL(window.location.href), state));} else {setState(next);}
  };
  window.addEventListener('popstate', restore);
  return () => window.removeEventListener('popstate', restore);
 }, [isDirty, state]);
 useEffect(() => {
  const warn = (event: BeforeUnloadEvent): void => {if (isDirty) {event.preventDefault(); event.returnValue = '';}};
  window.addEventListener('beforeunload', warn); return () => window.removeEventListener('beforeunload', warn);
 }, [isDirty]);
 useEffect(() => {
  if (wasOpen.current && !state.open) {entry.current?.focus();}
  if (hadDetail.current && !state.reference && state.open && row.current) {document.getElementById(`ai-result-${row.current}`)?.focus({preventScroll: true});}
  wasOpen.current = state.open; hadDetail.current = !!state.reference;
 }, [state.open, state.reference]);
 const navigate = (next: TResultLocation): void => {
  if (isDirty) {setPendingNavigation(next); return;}
  window.history.pushState(null, '', resultURL(new URL(window.location.href), next)); setState(next);
 };
 return <>
  <button ref={entry} type={'button'} className={'fixed right-4 bottom-20 z-[1000] cursor-pointer rounded-lg border border-(--color-border) bg-(--color-surface) px-4 py-2 text-sm shadow-md'} onClick={() => navigate({...state, open: true})}>{'AI Results'}</button>
  <DialogShell isOpen={state.open} onClose={() => navigate({...state, open: false, reference: null})} maxWidth={state.reference ? 'review' : 'md'} title={'AI Results'} subtitle={'Saved private analyses. Reviewing results does not change your photos.'}>
   {state.open && <div className={'p-4 text-sm [&_button]:cursor-pointer [&_button]:rounded [&_button]:border [&_button]:border-(--color-border) [&_button]:px-3 [&_button]:py-2 [&_button:disabled]:opacity-50 [&_label]:block [&_input]:w-full [&_input]:rounded [&_input]:border [&_input]:border-(--color-border) [&_input]:p-2 [&_select]:w-full [&_select]:rounded [&_select]:border [&_select]:border-(--color-border) [&_select]:p-2'}>
    {pendingNavigation && <div role={'alertdialog'} aria-label={'Unsaved draft edits'}>
     <p>{'Discard unsaved local edits or keep editing. Saved revisions remain available.'}</p>
     <button type={'button'} onClick={() => setPendingNavigation(null)}>{'Keep editing'}</button>
     <button type={'button'} onClick={() => {window.history.pushState(null, '', resultURL(new URL(window.location.href), pendingNavigation)); setState(pendingNavigation); setPendingNavigation(null); setIsDirty(false);}}>{'Discard unsaved edits'}</button>
                          </div>}
    <Description className={'sr-only'}>{'Browse saved analysis history, filter previous runs and inspect original results independently of the photo catalog.'}</Description>
    <div role={'region'} aria-label={'Results list'} hidden={!!state.reference} className={'max-h-[65dvh] space-y-3 overflow-y-auto'}>
     <ResultsFilters query={state.query} onApplyAction={query => navigate({open: true, query, reference: null})} />
     <ResultList
owner={owner} query={state.query} onPageAction={cursor => navigate({...state, query: {...state.query, cursor}})} onOpenAction={item => {
      row.current = item.id;
      navigate({...state, reference: item.analysisId ? {analysisId: item.analysisId} : {jobId: item.jobId, itemId: item.id}});
     }} />
    </div>
    {state.reference && <div className={'max-h-[65dvh] space-y-3 overflow-y-auto'}>
     <button autoFocus type={'button'} onClick={() => navigate({...state, reference: null})}>{'Back to results'}</button>
     <ResultDetail onVerified={onVerified} owner={owner} reference={state.reference} manualPendingIDs={manualPendingIDs} onDirtyChange={setIsDirty} />
                        </div>}
                  </div>}
  </DialogShell>
        </>;
}
export function AIResultsWorkspace(props: {owner: string; manualPendingIDs?: string[]; onVerified?: () => void}): ReactElement {return <ResultsSession key={props.owner} {...props} />;}
