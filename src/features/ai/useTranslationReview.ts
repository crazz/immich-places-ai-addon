import {useEffect, useRef, useState} from 'react';

import {fetchProviders} from './providerApi';
import {adoptTranslation, cancelTranslation, fetchTranslation, fetchTranslations, submitTranslation} from './translationApi';

import type {TDraft} from './draftTypes';
import type {TProviderList} from './providerApi';
import type {TTranslationPage, TTranslationRequest, TTranslationRun} from './translationTypes';

export type TTranslationRetry = {id: string; basis: string; basisKind: 'scene' | 'candidate'; languages: string[]};
type TState = {
 hasEdits: boolean; shouldConfirmClose: boolean; markEdited: () => void; closeReview: () => void; keepEditing: () => void; discardEdits: () => void;
 isOpen: boolean; setOpen: (value: boolean) => void; providers: TProviderList; error: string;
 run: TTranslationRun | null; isBusy: boolean; pending: TTranslationRequest | null;
 retry: TTranslationRetry | undefined; setRetry: (value: TTranslationRetry | undefined) => void;
 history: TTranslationPage; submit: (input: TTranslationRequest) => Promise<void>;
 adopt: (languages: string[]) => Promise<void>; update: (shouldCancel: boolean) => Promise<void>;
 browse: (before: string) => Promise<void>; selectRun: (id: string) => Promise<void>;
};
export function useTranslationReview({draft, disabled, onAdopt, onDirtyChange}: {draft: TDraft; disabled?: boolean; onAdopt: (value: TDraft) => void; onDirtyChange?: (value: boolean) => void}): TState {
 const [hasEdits,setEdits] = useState(false);const [shouldConfirmClose,setConfirmClose] = useState(false);
 function markEdited(): void {setEdits(true);onDirtyChange?.(true);}
 function closeReview(): void {if (hasEdits) {setConfirmClose(true);} else {setOpen(false);}}
 function keepEditing(): void {setConfirmClose(false);}
 function discardEdits(): void {setEdits(false);onDirtyChange?.(false);setConfirmClose(false);setOpen(false);}
 const [isOpen,setOpen] = useState(false);const [providers,setProviders] = useState<TProviderList>({enabled:false,items:[]});const [error,setError] = useState('');
 const [run,setRun] = useState<TTranslationRun | null>(null);const [isBusy,setBusy] = useState(false);const [pending,setPending] = useState<TTranslationRequest | null>(null);const controller = useRef<AbortController | null>(null);
 const [retry,setRetry] = useState<{id: string; basis: string; basisKind: 'scene' | 'candidate'; languages: string[]} | undefined>();
 const [history,setHistory] = useState<TTranslationPage>({ids:[],next:''});
 const requestVersion = useRef(0);
 useEffect(() => {
  if (!isOpen) {return;}const active = new AbortController();controller.current = active;setBusy(false);const version = ++requestVersion.current;
  void fetchProviders(active.signal).then(value => {if (!active.signal.aborted) {setProviders(value);}}).catch(() => {if (!active.signal.aborted) {setError('Translation providers unavailable. Reopen to try again.');}});
  void fetchTranslations(draft.id,'',active.signal).then(async page => {
   if (!active.signal.aborted && version === requestVersion.current) {setHistory(page);}
   if (page.ids[0]) {const value = await fetchTranslation(page.ids[0],active.signal);if (!active.signal.aborted && version === requestVersion.current) {setRun(value);}}
  }).catch(() => {if (!active.signal.aborted && version === requestVersion.current) {setError('Translation history unavailable. Reopen to reload saved history.');}});
  return () => active.abort();
 },[isOpen,draft.id,disabled]);
 async function perform<T>(request: (signal?: AbortSignal) => Promise<T>, apply: (value: T) => void, message: string): Promise<void> {
  const signal = controller.current?.signal;const version = ++requestVersion.current;setBusy(true);setError('');
  try {const value = await request(signal);if (!signal?.aborted && version === requestVersion.current) {apply(value);}}
  catch {if (!signal?.aborted && version === requestVersion.current) {setError(message);}}
  finally {if (!signal?.aborted && version === requestVersion.current) {setBusy(false);}}
 }
 async function submit(input: TTranslationRequest): Promise<void> {
  setPending(input);
  await perform(async signal => submitTranslation(input,signal), value => {setRun(value);setPending(null);setEdits(false);onDirtyChange?.(false);setHistory(page => ({...page,ids:[value.id,...page.ids.filter(id => id !== value.id)].slice(0,20)}));}, 'Submission not acknowledged. Recover this same submission before generating again.');
 }
 async function adopt(languages: string[]): Promise<void> {
  if (run) {await perform(async signal => adoptTranslation(run.id,draft.revision,languages,signal),onAdopt,'Adoption not acknowledged. Reload the saved draft to compare revisions and check active writes before trying again.');}
 }
 async function update(shouldCancel: boolean): Promise<void> {
  if (run) {await perform(async signal => shouldCancel ? cancelTranslation(run.id,signal) : fetchTranslation(run.id,signal),setRun,'Translation status unavailable. Refresh retained history before taking another action.');}
 }
 async function browse(before: string): Promise<void> {
  await perform(async signal => {const page = await fetchTranslations(draft.id,before,signal);const value = page.ids[0] ? await fetchTranslation(page.ids[0],signal) : null;return {page,value};},({page,value}) => {setHistory(page);setRun(value);},'History page unavailable. Try loading it again.');
 }
 async function selectRun(id: string): Promise<void> {
  await perform(async signal => fetchTranslation(id,signal),setRun,'Saved translation unavailable. Reload history.');
 }
 return {hasEdits,shouldConfirmClose,markEdited,closeReview,keepEditing,discardEdits,isOpen,setOpen,providers,error,run,isBusy,pending,retry,setRetry,history,submit,adopt,update,browse,selectRun};
}
