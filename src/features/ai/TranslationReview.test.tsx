import {act, fireEvent, render, screen} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {savedDraft} from './testing/draft';
import {translationRun} from './testing/translation';
import {TranslationReview} from './TranslationReview';

import type {TTranslationRun} from './translationTypes';

afterEach(() => vi.unstubAllGlobals());
const provider = {id: 'profile', revision: 1, name: 'Local provider', model: 'text-model', baseURL: 'https://provider.example/v1', enabled: true, hasSecret: true};

it('opens an explicit reviewed-text form with no automatic generation or hidden candidate prefill', async () => {
 const draft = savedDraft();draft.descriptions = [{language: 'en', text: 'Stale location claim', status: 'complete', basis: 'candidate', stale: true, factsRevision: 1, userSupplied: false}];
 const request = vi.fn<(path: string, init?: RequestInit) => Promise<Response>>(async path => new Response(JSON.stringify(path.endsWith('/providers') ? {enabled: true, items: [provider]} : {ids: [], next: ''})));
 vi.stubGlobal('fetch', request);
 render(<TranslationReview owner={'owner'} draft={draft} onAdopt={vi.fn()} />);
 fireEvent.click(screen.getByRole('button', {name: 'Translate reviewed text'}));
 expect(await screen.findByRole('option', {name: 'Local provider · text-model'})).toBeVisible();
 expect(screen.getByLabelText('Reviewed text basis')).toHaveValue('');
 expect(screen.getByRole('button', {name: 'Generate translations'})).toBeDisabled();
 expect(request.mock.calls.every(([, init]) => !init || init.method === 'GET' || init.method === undefined)).toBe(true);
});

it('explains that suggestions from an older factual basis are stale and cannot be adopted', async () => {
 const run = translationRun();vi.stubGlobal('fetch',vi.fn(async (path: string) => new Response(JSON.stringify(path.endsWith('/providers') ? {enabled:false,items:[]} : path.includes('?draftId=') ? {ids:[run.id],next:''} : run))));
 render(<TranslationReview owner={'owner'} draft={{...savedDraft(),revision:2,factsRevision:2}} onAdopt={vi.fn()} />);fireEvent.click(screen.getByRole('button',{name:'Translate reviewed text'}));
 expect(await screen.findByText(/These suggestions are stale/)).toBeVisible();expect(screen.getByRole('checkbox',{name:'Adopt en'})).toBeDisabled();
});

it('preserves unsaved basis edits until the owner explicitly discards or submits them', async () => {
 const dirty = vi.fn();vi.stubGlobal('fetch',vi.fn(async (path: string) => new Response(JSON.stringify(path.endsWith('/providers') ? {enabled:true,items:[provider]} : {ids:[],next:''}))));
 render(<TranslationReview owner={'owner'} draft={savedDraft()} onAdopt={vi.fn()} onDirtyChange={dirty} />);fireEvent.click(screen.getByRole('button',{name:'Translate reviewed text'}));await screen.findByRole('option',{name:'Local provider · text-model'});
 fireEvent.change(screen.getByLabelText('Reviewed text basis'),{target:{value:'Unsaved reviewed facts.'}});expect(dirty).toHaveBeenLastCalledWith(true);
 fireEvent.click(screen.getByRole('button',{name:'Close translation review'}));fireEvent.click(screen.getByRole('button',{name:'Keep translation edits'}));expect(screen.getByLabelText('Reviewed text basis')).toHaveValue('Unsaved reviewed facts.');
 fireEvent.click(screen.getByRole('button',{name:'Close translation review'}));fireEvent.click(screen.getByRole('button',{name:'Discard translation edits'}));expect(dirty).toHaveBeenLastCalledWith(false);expect(screen.queryByLabelText('Reviewed text basis')).not.toBeInTheDocument();
});

it('keeps a newly submitted run when an older history read arrives late', async () => {
 const old = translationRun();const current = {...old,id:'77777777-7777-4777-8777-777777777777',items:[{...old.items[0],text:'Newer suggestion.'},old.items[1]]};let reply: ((value: Response) => void) | undefined;
 vi.stubGlobal('fetch',vi.fn(async (path: string,init?: RequestInit) => {
  if (init?.method === 'POST') {return new Response(JSON.stringify(current));}
  if (path.endsWith(old.id)) {return new Promise<Response>(resolve => {reply = resolve;});}
  return new Response(JSON.stringify(path.endsWith('/providers') ? {enabled:true,items:[provider]} : {ids:[old.id],next:''}));
 }));
 render(<TranslationReview owner={'owner'} draft={savedDraft()} onAdopt={vi.fn()} />);fireEvent.click(screen.getByRole('button',{name:'Translate reviewed text'}));await vi.waitFor(() => expect(reply).toBeDefined());
 fireEvent.change(screen.getByLabelText('Reviewed text basis'),{target:{value:old.request.basis}});fireEvent.change(screen.getByLabelText('Translation provider'),{target:{value:'profile'}});fireEvent.click(screen.getByRole('checkbox',{name:/I approve sending/}));fireEvent.click(screen.getByRole('button',{name:'Generate translations'}));await screen.findByText('Newer suggestion.');
 await act(async () => reply?.(new Response(JSON.stringify(old))));expect(screen.getByText('Newer suggestion.')).toBeVisible();
});

it('recovers a lost submission acknowledgement with the identical approved request', async () => {
 const run = translationRun();let submissions = 0;
 const request = vi.fn(async (path: string, init?: RequestInit) => {
  if (init?.method === 'POST') {submissions++;if (submissions === 1) {throw new Error('lost acknowledgement');}return new Response(JSON.stringify(run));}
  return new Response(JSON.stringify(path.endsWith('/providers') ? {enabled:true,items:[provider]} : {ids:[],next:''}));
 });
 vi.stubGlobal('fetch',request);render(<TranslationReview owner={'owner'} draft={savedDraft()} onAdopt={vi.fn()} />);fireEvent.click(screen.getByRole('button',{name:'Translate reviewed text'}));await screen.findByRole('option',{name:'Local provider · text-model'});
 fireEvent.change(screen.getByLabelText('Reviewed text basis'),{target:{value:run.request.basis}});fireEvent.change(screen.getByLabelText('Translation provider'),{target:{value:'profile'}});fireEvent.click(screen.getByRole('checkbox',{name:/I approve sending/}));fireEvent.click(screen.getByRole('button',{name:'Generate translations'}));
 fireEvent.click(await screen.findByRole('button',{name:'Recover submitted translation'}));await screen.findByText('en: complete');
 const writes = request.mock.calls.filter(([,init]) => init?.method === 'POST');expect(writes).toHaveLength(2);expect(writes[0][1]?.body).toBe(writes[1][1]?.body);
});

it('browses bounded older history without generating or adopting anything', async () => {
 const run = translationRun();const old = {...run,id:'77777777-7777-4777-8777-777777777777',request:{...run.request,basis:'Earlier reviewed scene.'}};
 const request = vi.fn<(path: string, init?: RequestInit) => Promise<Response>>(async path => new Response(JSON.stringify(path.endsWith('/providers') ? {enabled:false,items:[]} : path.includes(`before=${run.id}`) ? {ids:[old.id],next:''} : path.includes('?draftId=') ? {ids:[run.id],next:run.id} : path.endsWith(old.id) ? old : run)));
 vi.stubGlobal('fetch',request);render(<TranslationReview owner={'owner'} draft={savedDraft()} onAdopt={vi.fn()} />);fireEvent.click(screen.getByRole('button',{name:'Translate reviewed text'}));await screen.findByText('en: complete');
 fireEvent.click(screen.getByRole('button',{name:'Older translation runs'}));expect(await screen.findByText('Earlier reviewed scene.')).toBeVisible();
 fireEvent.click(screen.getByRole('button',{name:'Newest translation runs'}));await vi.waitFor(() => expect(screen.queryByText('Earlier reviewed scene.')).not.toBeInTheDocument());
 expect(request.mock.calls.every(([,init]) => init?.method !== 'POST')).toBe(true);
});

it('fences late adoption after an account, draft or unsaved-editor change', async () => {
 for (const change of ['owner','draft','dirty']) {
  const run = translationRun();const draft = savedDraft();const adopted = vi.fn();let reply: (value: Response) => void = () => {};
  vi.stubGlobal('fetch',vi.fn(async (path: string) => {
   if (path.endsWith('/adopt')) {return new Promise<Response>(resolve => {reply = resolve;});}
   return new Response(JSON.stringify(path.endsWith('/providers') ? {enabled:true,items:[provider]} : path.includes('?draftId=') ? {ids:[run.id],next:''} : run));
  }));
  const view = render(<TranslationReview owner={'owner'} draft={draft} onAdopt={adopted} />);
  fireEvent.click(screen.getByRole('button',{name:'Translate reviewed text'}));await screen.findByText('en: complete');
  fireEvent.click(screen.getByRole('checkbox',{name:'Adopt en'}));fireEvent.click(screen.getByRole('button',{name:'Adopt selected translations'}));
  view.rerender(<TranslationReview owner={change === 'owner' ? 'other' : 'owner'} draft={change === 'draft' ? {...draft,id:'66666666-6666-4666-8666-666666666666'} : draft} disabled={change === 'dirty'} onAdopt={adopted} />);
  await act(async () => reply(new Response(JSON.stringify({...draft,revision:2}))));
  expect(adopted).not.toHaveBeenCalled();view.unmount();
 }
});

it('refreshes queued progress with reads and cancels unfinished languages only on an explicit action', async () => {
 const run = translationRun();let reads = 0;
 const queued: TTranslationRun = {...run, items: run.items.map(item => ({...item, text: null, state: 'queued', failure: ''}))};
 const request = vi.fn<(path: string, init?: RequestInit) => Promise<Response>>(async path => {
  if (path.endsWith('/providers')) {return new Response(JSON.stringify({enabled: true, items: [provider]}));}
  if (path.includes('?draftId=')) {return new Response(JSON.stringify({ids: [run.id], next: ''}));}
  if (path.endsWith('/cancel')) {return new Response(JSON.stringify({...run, items: [run.items[0], {...queued.items[1], state: 'canceled'}]}));}
  reads++;return new Response(JSON.stringify(reads === 1 ? queued : {...run, items: [run.items[0], queued.items[1]]}));
 });
 vi.stubGlobal('fetch', request);render(<TranslationReview owner={'owner'} draft={savedDraft()} onAdopt={vi.fn()} />);
 fireEvent.click(screen.getByRole('button', {name: 'Translate reviewed text'}));await screen.findByText('en: queued');
 fireEvent.click(screen.getByRole('button', {name: 'Refresh translations'}));await screen.findByText('en: complete');
 expect(request.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(false);
 fireEvent.click(screen.getByRole('button', {name: 'Cancel unfinished translations'}));expect(await screen.findByText('uk: canceled')).toBeVisible();expect(screen.getByText('en: complete')).toBeVisible();
 expect(request.mock.calls.filter(([, init]) => init?.method === 'POST').map(([path]) => path)).toEqual([expect.stringContaining(`/${run.id}/cancel`)]);
});

it('requires a fresh review before retrying only selected unsuccessful languages', async () => {
 const run = translationRun();const retry = {...run, id: '77777777-7777-4777-8777-777777777777', request: {...run.request, parentId: run.id, languages: ['uk']}, items: [run.items[1]]};
 const request = vi.fn(async (path: string, init?: RequestInit) => new Response(JSON.stringify(path.endsWith('/providers') ? {enabled: true, items: [provider]} : path.includes('?draftId=') ? {ids: [run.id], next: ''} : init?.method === 'POST' ? retry : run)));
 vi.stubGlobal('fetch', request);render(<TranslationReview owner={'owner'} draft={savedDraft()} onAdopt={vi.fn()} />);
 fireEvent.click(screen.getByRole('button', {name: 'Translate reviewed text'}));await screen.findByText('uk: failed');
 fireEvent.click(screen.getByRole('checkbox', {name: 'Retry uk'}));fireEvent.click(screen.getByRole('button', {name: 'Review retry languages'}));
 expect(screen.getByLabelText('Translation languages')).toHaveValue('uk');expect(screen.getByRole('button', {name: 'Generate translations'})).toBeDisabled();
 expect(request.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(false);
 fireEvent.change(screen.getByLabelText('Translation provider'), {target: {value: 'profile'}});fireEvent.click(screen.getByRole('checkbox', {name: /I approve sending/}));fireEvent.click(screen.getByRole('button', {name: 'Generate translations'}));
 await vi.waitFor(() => expect(request.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(true));
 const write = request.mock.calls.find(([, init]) => init?.method === 'POST');expect(JSON.parse(String(write?.[1]?.body))).toEqual({...retry.request, key: expect.any(String)});
});

it('reads retained history while disabled and adopts only an explicitly selected successful language', async () => {
 const run = translationRun();const draft = savedDraft();const adopted = vi.fn();
 const updated = {...draft, revision: 2};
 const request = vi.fn<(path: string, init?: RequestInit) => Promise<Response>>(async path => new Response(JSON.stringify(path.endsWith('/providers') ? {enabled: false, items: [provider]} : path.endsWith('/adopt') ? updated : path.includes('?draftId=') ? {ids: [run.id], next: ''} : run)));
 vi.stubGlobal('fetch', request);render(<TranslationReview owner={'owner'} draft={draft} onAdopt={adopted} />);
 fireEvent.click(screen.getByRole('button', {name: 'Translate reviewed text'}));
 expect(await screen.findByText('en: complete')).toBeVisible();
 expect(screen.getByRole('button', {name: 'Adopt selected translations'})).toBeDisabled();
 fireEvent.click(screen.getByRole('checkbox', {name: 'Adopt en'}));
 fireEvent.click(screen.getByRole('button', {name: 'Adopt selected translations'}));
 await vi.waitFor(() => expect(adopted).toHaveBeenCalledWith(updated));
 const writes = request.mock.calls.filter(([, init]) => init?.method === 'POST');expect(writes).toHaveLength(1);expect(writes[0][0]).toContain(`/${run.id}/adopt`);expect(JSON.parse(String(writes[0][1]?.body))).toEqual({revision: draft.revision, languages: ['en']});
 expect(screen.getByRole('checkbox', {name: 'Adopt uk'})).toBeDisabled();
});

it('sends exactly the reviewed basis and languages and retains partial suggestions without editing the draft', async () => {
 const run = translationRun(); const adopted = vi.fn();
 const request = vi.fn(async (path: string, init?: RequestInit) => new Response(JSON.stringify(path.endsWith('/providers') ? {enabled: true, items: [provider]} : init?.method === 'POST' ? run : {ids: [], next: ''})));
 vi.stubGlobal('fetch', request);
 render(<TranslationReview owner={'owner'} draft={savedDraft()} onAdopt={adopted} />);
 fireEvent.click(screen.getByRole('button', {name: 'Translate reviewed text'}));
 await screen.findByRole('option', {name: 'Local provider · text-model'});
 fireEvent.change(screen.getByLabelText('Reviewed text basis'), {target: {value: run.request.basis}});
 fireEvent.change(screen.getByLabelText('Translation provider'), {target: {value: 'profile'}});
 fireEvent.click(screen.getByRole('checkbox', {name: /I approve sending/}));
 fireEvent.click(screen.getByRole('button', {name: 'Generate translations'}));
 expect(await screen.findByText('en: complete')).toBeVisible();expect(screen.getByText('uk: failed')).toBeVisible();
 const calls = request.mock.calls.filter(([, init]) => init?.method === 'POST');expect(calls).toHaveLength(1);
 expect(JSON.parse(String(calls[0][1]?.body))).toEqual({...run.request, key: expect.any(String)});
 expect(adopted).not.toHaveBeenCalled();
});
