import {render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {expect, it, vi} from 'vitest';

import {DraftEditor} from './DraftEditor';
import {descriptionDraft} from './testing/draft';

it('M02 deliberately selects the analyzed-photo mirror alongside a standard decision', async () => {
 const user = userEvent.setup(); const save = vi.fn().mockResolvedValue(undefined);
 const draft = {...descriptionDraft(), fields: ['description' as const], primaryLanguage: 'uk', descriptionPolicy: 'replace' as const};
 render(<DraftEditor draft={draft} isBusy={false} onSave={save} />);
 const choice = screen.getByRole('checkbox', {name: 'Mirror optional metadata for this photo'});
 expect(choice).not.toBeChecked(); choice.focus(); await user.keyboard(' ');
 await user.click(screen.getByRole('checkbox', {name: 'Mirror description uk'}));
 await user.click(screen.getByRole('button', {name: 'Stage selected fields'}));
 await waitFor(() => expect(save).toHaveBeenCalledWith({camera: null, fields: ['description'], mirror: {direction: false, precision: false, place: false, languages: ['uk'], provenance: false, model: false}, state: 'staged'}));
 expect(screen.getByText(`Metadata target: ${draft.assetId}`)).toBeVisible();
 expect(screen.getByText(/custom metadata does not promise native Immich display or file\/EXIF updates/i)).toBeVisible();
 expect(screen.getByLabelText('Description uk')).toHaveValue('  Місто\n河  ');
});

it('M04 independently reviews stale precision and direction while excluding unreviewed languages', async () => {
 const user = userEvent.setup(); const save = vi.fn().mockResolvedValue(undefined); const draft = {...descriptionDraft(), camera: {latitude: 0, longitude: 12}, fields: ['gps' as const], radius: 5000, radiusStale: true, heading: 90, headingStale: true};
 draft.descriptions.push({...draft.descriptions[0], language: 'en', basis: 'candidate', stale: true, text: 'Old place'});
 render(<DraftEditor draft={draft} isBusy={false} onSave={save} />);
 await user.click(screen.getByLabelText('Mirror optional metadata for this photo'));
 expect(screen.getByLabelText('Mirror direction')).toBeDisabled(); expect(screen.getByLabelText('Mirror precision')).toBeDisabled(); expect(screen.getByLabelText('Mirror description en')).toBeDisabled();
 await user.click(screen.getByLabelText('Review heading for this camera point'));
 expect(screen.getByLabelText('Mirror direction')).toBeEnabled(); expect(screen.getByLabelText('Mirror precision')).toBeDisabled();
 await user.click(screen.getByLabelText('Review precision for this camera point'));
 await user.click(screen.getByLabelText('Mirror precision')); await user.click(screen.getByLabelText('Mirror direction')); await user.click(screen.getByLabelText('Mirror description uk'));
 expect(screen.getByLabelText('Include model name')).toBeDisabled(); await user.click(screen.getByLabelText('Mirror minimal provenance')); await user.click(screen.getByLabelText('Include model name'));
 await user.click(screen.getByRole('button', {name: 'Stage GPS draft'}));
 await waitFor(() => expect(save).toHaveBeenCalledWith({camera: {latitude: 0, longitude: 12}, fields: ['gps'], reviewHeading: true, reviewPrecision: true, mirror: {direction: true, precision: true, place: false, languages: ['uk'], provenance: true, model: true}, state: 'staged'}));
 expect(screen.getByLabelText('Description en')).toHaveValue('Old place');
 expect(screen.getByLabelText('Mirror description en')).not.toBeChecked();
});

it('M03 prevents empty selection and mirror-only staging without implicit probes', async () => {
 const user = userEvent.setup(); const save = vi.fn(); const request = vi.spyOn(globalThis, 'fetch'); const draft = {...descriptionDraft(), camera: {latitude: 0, longitude: 12}, fields: ['gps' as const]};
 render(<DraftEditor draft={draft} isBusy={false} onSave={save} />);
 await user.click(screen.getByLabelText('Mirror optional metadata for this photo'));
 expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeDisabled();
 await user.click(screen.getByLabelText('Mirror description uk'));
 expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeEnabled();
 await user.click(screen.getByLabelText('Select GPS for this photo'));
 expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeDisabled();
 expect(save).not.toHaveBeenCalled(); expect(request).not.toHaveBeenCalled(); request.mockRestore();
});

it('M05 blocks stale selected content and a ninth language without truncating local records', async () => {
 const user = userEvent.setup(); const save = vi.fn(); const draft = {...descriptionDraft(), camera: {latitude: 0, longitude: 12}, fields: ['gps' as const], radiusStale: true, mirror: {direction: false, precision: true, place: false, languages: [], provenance: false, model: false}};
 draft.descriptions = ['en', 'uk', 'fr', 'de', 'it', 'es', 'pt', 'ja', 'zh'].map(language => ({...draft.descriptions[0], language}));
 render(<DraftEditor draft={draft} isBusy={false} onSave={save} />);
 expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeDisabled();
 await user.click(screen.getByLabelText('Review precision for this camera point'));
 for (const item of draft.descriptions.slice(0, 8)) {await user.click(screen.getByLabelText(`Mirror description ${item.language}`));}
 expect(screen.getByLabelText('Mirror description zh')).toBeDisabled(); expect(screen.getByLabelText('Description zh')).toHaveValue('  Місто\n河  ');
 await user.click(screen.getByLabelText('Mirror optional metadata for this photo'));
 expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeEnabled();
 await user.click(screen.getByRole('button', {name: 'Save draft'}));
 await waitFor(() => expect(save).toHaveBeenCalledWith({camera: {latitude: 0, longitude: 12}, fields: ['gps'], reviewPrecision: true, mirror: null}));
});

it('selects the immutable chosen place explicitly and normalizes a nil language list for editing', async () => {
 const user = userEvent.setup(); const save = vi.fn(); const draft = {...descriptionDraft(), candidateId: 'place-1', camera: {latitude: 0, longitude: 12}, fields: ['gps' as const], mirror: {direction: false, precision: false, place: false, languages: null, provenance: true, model: false}};
 render(<DraftEditor draft={draft} candidates={[{id: 'place-1', name: 'Київ · riverside', camera: draft.camera}]} isBusy={false} onSave={save} />);
 expect(screen.getByLabelText('Mirror description uk')).not.toBeChecked(); expect(screen.getByLabelText('Mirror chosen place')).not.toBeChecked();
 await user.click(screen.getByLabelText('Mirror chosen place')); await user.click(screen.getByLabelText('Include model name')); await user.click(screen.getByRole('button', {name: 'Stage GPS draft'}));
 await waitFor(() => expect(save).toHaveBeenCalledWith({camera: draft.camera, fields: ['gps'], mirror: {...draft.mirror, languages: [], place: true, model: true}, state: 'staged'}));
 expect(screen.getByText('Chosen place for mirror: Київ · riverside')).toBeVisible();
});

it('invalidates precision review after a new camera correction and permits removing stale selected precision', async () => {
 const {fireEvent} = await import('@testing-library/react'); const user = userEvent.setup(); const save = vi.fn(); const draft = {...descriptionDraft(), camera: {latitude: 0, longitude: 12}, radius: 5000, radiusStale: true, fields: ['gps' as const], mirror: {direction: false, precision: true, place: false, languages: ['uk'], provenance: false, model: false}};
 render(<DraftEditor draft={draft} isBusy={false} onSave={save} />);
 await user.click(screen.getByLabelText('Review precision for this camera point')); expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeEnabled();
 fireEvent.change(screen.getByLabelText('Camera latitude'), {target: {value: '1'}});
 expect(screen.getByLabelText('Review precision for this camera point')).not.toBeChecked(); expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeDisabled();
 await user.click(screen.getByLabelText('Mirror precision')); expect(screen.getByLabelText('Mirror precision')).not.toBeChecked(); expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeEnabled();
 await user.click(screen.getByRole('button', {name: 'Stage GPS draft'})); await waitFor(() => expect(save).toHaveBeenCalledWith({camera: {latitude: 1, longitude: 12}, fields: ['gps'], mirror: {...draft.mirror, precision: false}, state: 'staged'}));
});

it('removes stale selected languages and direction after a later camera edit invalidates heading review', async () => {
 const {fireEvent} = await import('@testing-library/react'); const user = userEvent.setup(); const draft = {...descriptionDraft(), camera: {latitude: 0, longitude: 12}, fields: ['gps' as const], heading: 90, headingStale: true, mirror: {direction: true, precision: false, place: false, languages: ['en', 'uk'], provenance: false, model: false}};
 draft.descriptions.push({...draft.descriptions[0], language: 'en', basis: 'candidate', stale: true});
 render(<DraftEditor draft={draft} isBusy={false} onSave={vi.fn()} />);
 await user.click(screen.getByLabelText('Mirror description en')); expect(screen.getByLabelText('Mirror description en')).not.toBeChecked();
 await user.click(screen.getByLabelText('Review heading for this camera point')); expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeEnabled();
 fireEvent.change(screen.getByLabelText('Camera longitude'), {target: {value: '13'}}); expect(screen.getByLabelText('Review heading for this camera point')).not.toBeChecked(); expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeDisabled();
 await user.click(screen.getByLabelText('Mirror direction')); expect(screen.getByLabelText('Mirror direction')).not.toBeChecked(); expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeEnabled(); expect(screen.getByLabelText('Description en')).toHaveValue(draft.descriptions[1].text);
});
