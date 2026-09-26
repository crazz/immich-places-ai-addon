import {render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {expect, it, vi} from 'vitest';

import {DraftEditor} from './DraftEditor';
import {savedDraft} from './testing/draft';

it('selects one primary description with deliberate replacement independently of GPS', async () => {
 const user = userEvent.setup(); const draft = savedDraft(); const save = vi.fn().mockResolvedValue(undefined);
 draft.descriptions = [{language: 'uk', status: 'complete', text: '  Місто\n河  ', basis: 'scene_only', stale: false, factsRevision: 1, userSupplied: false}];
 render(<DraftEditor draft={draft} isBusy={false} hasManualConflict onSave={save} />);
 expect(screen.getByLabelText('Description policy')).toHaveValue('preserve');
 expect(screen.getByLabelText('Select description for this photo')).not.toBeChecked();
 await user.selectOptions(screen.getByLabelText('Primary description language'), 'uk');
 await user.selectOptions(screen.getByLabelText('Description policy'), 'replace');
 await user.click(screen.getByLabelText('Select description for this photo'));
 await user.click(screen.getByRole('button', {name: 'Stage selected fields'}));
 await waitFor(() => expect(save).toHaveBeenCalledWith({camera: null, fields: ['description'], primaryLanguage: 'uk', descriptionPolicy: 'replace', state: 'staged'}));
 expect(screen.getByLabelText('Select GPS for this photo')).not.toBeChecked();
 expect(screen.getByLabelText('Description uk')).toHaveValue('  Місто\n河  ');
});

it('C02 preserves existing description during GPS-only staging with unavailable and stale text', async () => {
 const user = userEvent.setup(); const save = vi.fn().mockResolvedValue(undefined); const draft = {...savedDraft(), camera: {latitude: 0, longitude: 12}};
 draft.descriptions = [{language: 'en', status: 'complete', text: 'Stale', basis: 'candidate', stale: true, factsRevision: 1, userSupplied: false}, {language: 'uk', status: 'unavailable', text: null, basis: 'candidate', stale: true, factsRevision: 1, userSupplied: false}];
 draft.baseline.description = {presence: 'value', value: 'Existing text'};
 render(<DraftEditor draft={draft} isBusy={false} onSave={save} />);
 await user.click(screen.getByLabelText('Select GPS for this photo'));
 await user.click(screen.getByRole('button', {name: 'Stage GPS draft'}));
 await waitFor(() => expect(save).toHaveBeenCalledWith({camera: {latitude: 0, longitude: 12}, fields: ['gps'], state: 'staged'}));
 expect(screen.getByLabelText('Description policy')).toHaveValue('preserve');
 expect(screen.getByLabelText('Description en')).toHaveValue('Stale');
 expect(draft.baseline.description.value).toBe('Existing text');
});

it('blocks stale or empty description selection until explicit current review and rejects no-field staging', async () => {
 const user = userEvent.setup(); const save = vi.fn(); const draft = savedDraft();
 draft.descriptions = [{language: 'uk', status: 'complete', text: 'Old text', basis: 'candidate', stale: true, factsRevision: 1, userSupplied: false}];
 render(<DraftEditor draft={draft} isBusy={false} onSave={save} />);
 expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeDisabled();
 await user.selectOptions(screen.getByLabelText('Primary description language'), 'uk');
 await user.selectOptions(screen.getByLabelText('Description policy'), 'replace');
 expect(screen.getByLabelText('Select description for this photo')).toBeDisabled();
 await user.click(screen.getByLabelText('Review uk for this camera point'));
 await user.click(screen.getByLabelText('Select description for this photo'));
 expect(screen.getByRole('button', {name: 'Stage selected fields'})).toBeEnabled();
 await user.clear(screen.getByLabelText('Description uk'));
 expect(screen.getByRole('button', {name: 'Stage selected fields'})).toBeDisabled();
 expect(screen.getByLabelText('Description uk')).toHaveValue('');
 expect(save).not.toHaveBeenCalled();
});

it('requires coordinates only while GPS is selected and omits description again when preserve is chosen', async () => {
 const user = userEvent.setup(); const save = vi.fn().mockResolvedValue(undefined); const draft = savedDraft();
 draft.descriptions = [{language: 'uk', status: 'complete', text: 'Scene', basis: 'scene_only', stale: false, factsRevision: 1, userSupplied: false}];
 render(<DraftEditor draft={draft} isBusy={false} onSave={save} />);
 await user.selectOptions(screen.getByLabelText('Primary description language'), 'uk');
 await user.selectOptions(screen.getByLabelText('Description policy'), 'replace');
 await user.click(screen.getByLabelText('Select description for this photo'));
 await user.click(screen.getByLabelText('Select GPS for this photo'));
 expect(screen.getByRole('button', {name: 'Stage selected fields'})).toBeDisabled();
 await user.click(screen.getByLabelText('Select GPS for this photo'));
 expect(screen.getByRole('button', {name: 'Stage selected fields'})).toBeEnabled();
 await user.selectOptions(screen.getByLabelText('Description policy'), 'preserve');
 expect(screen.getByLabelText('Select description for this photo')).not.toBeChecked();
 expect(screen.getByRole('button', {name: 'Stage GPS draft'})).toBeDisabled();
 await user.click(screen.getByRole('button', {name: 'Save draft'}));
 await waitFor(() => expect(save).toHaveBeenCalledWith({camera: null, fields: [], primaryLanguage: 'uk'}));
});
