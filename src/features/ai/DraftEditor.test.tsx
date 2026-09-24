import {act, fireEvent, render, screen, waitFor} from '@testing-library/react';
import L from 'leaflet';
import {expect, it, vi} from 'vitest';

import * as mapBase from '@/features/map';

import {DraftEditor} from './DraftEditor';
import {savedDraft} from './testing/draft';

it('stages only GPS while keeping stale languages and heading visible and independently editable', async () => {
 const draft = {...savedDraft(), camera: {latitude: 1, longitude: 2}, heading: 40, headingStale: true, radius: 5000, radiusStale: true};
 draft.descriptions = [{language: 'en', status: 'complete', text: 'Old text', basis: 'candidate', stale: true, factsRevision: 1, userSupplied: false}, {language: 'uk', status: 'unavailable', text: null, basis: 'candidate', stale: true, factsRevision: 1, userSupplied: false}];
 const save = vi.fn().mockResolvedValue(undefined);
 render(<DraftEditor draft={draft} isBusy={false} onSave={save} />);
 expect(screen.getByText(/Radius: 5000 m.*needs review/)).toBeVisible();
 expect(screen.getByText('uk: unavailable · needs review')).toBeVisible();
 fireEvent.change(screen.getByLabelText('Heading degrees'), {target: {value: '0'}});
 fireEvent.change(screen.getByLabelText('Description en'), {target: {value: 'User corrected'}});
 fireEvent.click(screen.getByLabelText('Select GPS for this photo'));
 fireEvent.click(screen.getByRole('button', {name: 'Stage GPS draft'}));
 await waitFor(() => expect(save).toHaveBeenCalledWith({camera: {latitude: 1, longitude: 2}, fields: ['gps'], heading: 0, descriptions: {en: {text: 'User corrected'}}, state: 'staged'}));
});

it('blocks GPS staging until overlapping manual work is explicitly resolved', () => {
 const draft = {...savedDraft(), camera: {latitude: 1, longitude: 2}, fields: ['gps'] as ('gps')[]};
 const save = vi.fn();
 const view = render(<DraftEditor draft={draft} isBusy={false} hasManualConflict onSave={save} />);
 expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeDisabled();
 expect(screen.getByText(/A manual location is pending for this photo/)).toBeVisible();
 expect(screen.getByRole('button', {name: 'Save draft'})).toBeEnabled();
 view.rerender(<DraftEditor draft={draft} isBusy={false} hasManualConflict={false} onSave={save} />);
 expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeEnabled();
 expect(save).not.toHaveBeenCalled();
});

it('edits an isolated map locally and preserves numeric access when tiles fail', async () => {
 const factory = vi.spyOn(mapBase, 'createBaseMap'); const save = vi.fn();
 const view = render(<DraftEditor draft={{...savedDraft(), camera: {latitude: 1, longitude: 2}}} isBusy={false} onSave={save} />);
 expect(await screen.findByRole('region', {name: 'Editable draft map'})).toBeVisible();
 await waitFor(() => expect(factory).toHaveBeenCalled());
 const {map, tiles} = factory.mock.results[0].value;
 act(() => {map.fire('click', {latlng: L.latLng(0, 0)}); tiles.fire('tileerror');});
 expect(screen.getByLabelText('Camera latitude')).toHaveValue(0);
 expect(screen.getByLabelText('Camera longitude')).toHaveValue(0);
 expect(screen.getByText(/Draft map tiles unavailable/)).toBeVisible();
 expect(save).not.toHaveBeenCalled();
 const remove = vi.spyOn(map, 'remove'); view.unmount(); expect(remove).toHaveBeenCalled(); factory.mockRestore();
});

it('saves an explicit proposal candidate choice separately from inspection', async () => {
 const save = vi.fn().mockResolvedValue(undefined);
 const candidates = [{id: 'candidate-1', name: 'Coarse candidate', camera: {latitude: 50, longitude: 14}}];
 render(<DraftEditor draft={savedDraft()} candidates={candidates} isBusy={false} onSave={save} />);
 fireEvent.change(screen.getByLabelText('Draft camera candidate'), {target: {value: 'candidate-1'}});
 expect(screen.getByLabelText('Camera latitude')).toHaveValue(50);
 expect(save).not.toHaveBeenCalled();
 fireEvent.click(screen.getByRole('button', {name: 'Save draft'}));
 await waitFor(() => expect(save).toHaveBeenCalledWith({candidateId: 'candidate-1', fields: []}));
});

it('prevents edits to controls while a save is pending', () => {
 render(<DraftEditor draft={savedDraft()} isBusy onSave={vi.fn()} />);
 expect(screen.getByLabelText('Camera latitude')).toBeDisabled();
 expect(screen.getByLabelText('Heading degrees')).toBeDisabled();
 expect(screen.getByLabelText('Select GPS for this photo')).toBeDisabled();
});

it('shows inherited estimates as stale while camera corrections are still unsaved', () => {
 render(<DraftEditor draft={{...savedDraft(),camera:{latitude:1,longitude:2},radius:5000,radiusBasis:'model_estimate',heading:40}} isBusy={false} onSave={vi.fn()} />);
 fireEvent.change(screen.getByLabelText('Camera latitude'),{target:{value:'3'}});
 expect(screen.getByText(/Radius: 5000 m.*needs review/)).toBeVisible();
 expect(screen.getByText('Heading: needs review')).toBeVisible();
});

it('retains corrected language text when its optional review checkbox is cleared', () => {
 const draft=savedDraft();draft.descriptions=[{language:'en',status:'complete',text:'Original',basis:'candidate',stale:true,factsRevision:1,userSupplied:false}];
 render(<DraftEditor draft={draft} isBusy={false} onSave={vi.fn()} />);
 fireEvent.change(screen.getByLabelText('Description en'),{target:{value:'Correction'}});
 fireEvent.click(screen.getByLabelText('Review en for this camera point'));
 fireEvent.click(screen.getByLabelText('Review en for this camera point'));
 expect(screen.getByLabelText('Description en')).toHaveValue('Correction');
});
