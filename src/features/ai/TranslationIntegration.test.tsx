import {fireEvent, render, screen} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {DraftReview} from './DraftReview';
import {parseReview} from './reviewParser';
import {savedDraft} from './testing/draft';
import {resultDetail} from './testing/resultDetail';

afterEach(() => vi.unstubAllGlobals());
it('makes translation review available beside the saved local draft without starting provider work', async () => {
 const draft = savedDraft();const detail = resultDetail();detail.entry.draftId = draft.id;
 const fetch = vi.fn<(path: string, init?: RequestInit) => Promise<Response>>(async path => new Response(JSON.stringify(path.includes('write-operations') ? {items:[],nextCursor:null} : draft)));vi.stubGlobal('fetch',fetch);
 const review = parseReview(detail);if (!review) {throw new Error('Invalid synthetic review');}
 render(<DraftReview owner={'owner'} detail={detail} review={review} />);
 expect(await screen.findByRole('button',{name:'Translate reviewed text'})).toBeVisible();expect(fetch.mock.calls.every(([path,init]) => !path.includes('/ai/translations') && init?.method !== 'POST')).toBe(true);
});

it('includes translation edits in the existing navigation guard without disabling explicit generation', async () => {
 const draft = savedDraft();const detail = resultDetail();detail.entry.draftId = draft.id;const dirty = vi.fn();
 vi.stubGlobal('fetch',vi.fn(async (path: string) => new Response(JSON.stringify(path.endsWith('/providers') ? {enabled:true,items:[{id:'profile',revision:1,name:'Provider',model:'text',baseURL:'https://provider.example/v1',enabled:true,hasSecret:true}]} : path.includes('/translations') ? {ids:[],next:''} : path.includes('write-operations') ? {items:[],nextCursor:null} : draft))));
 const review = parseReview(detail);if (!review) {throw new Error('Invalid synthetic review');}
 render(<DraftReview owner={'owner'} detail={detail} review={review} onDirtyChange={dirty} />);
 fireEvent.click(await screen.findByRole('button',{name:'Translate reviewed text'}));await screen.findByRole('option',{name:'Provider · text'});
 fireEvent.change(screen.getByLabelText('Reviewed text basis'),{target:{value:'Pending reviewed facts.'}});await vi.waitFor(() => expect(dirty).toHaveBeenLastCalledWith(true));
 fireEvent.change(screen.getByLabelText('Translation provider'),{target:{value:'profile'}});fireEvent.click(screen.getByRole('checkbox',{name:/I approve sending/}));expect(screen.getByRole('button',{name:'Generate translations'})).toBeEnabled();
});
