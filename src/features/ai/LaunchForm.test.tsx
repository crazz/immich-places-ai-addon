import {render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import {submitJob} from './jobApi';
import {LaunchForm} from './LaunchForm';
import {fetchProviders} from './providerApi';
import {aiPreview, aiProfile, aiProgress} from './testing/fixtures';

vi.mock('./providerApi', () => ({fetchProviders: vi.fn()}));
vi.mock('./jobApi', () => ({submitJob: vi.fn()}));
afterEach(() => vi.resetAllMocks());

it('requires fresh image consent after changing choices and submits one exact Visual run', async () => {
 const user = userEvent.setup();
 vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [aiProfile]});
 vi.mocked(submitJob).mockResolvedValue(aiProgress());
 const onSubmitted = vi.fn();
 render(<LaunchForm preview={aiPreview()} onSubmittedAction={onSubmitted} />);
 const consent = await screen.findByLabelText(/I consent to sending/);
 expect(consent).not.toBeChecked();
 expect(screen.getByRole('button', {name: 'Start analysis'})).toBeDisabled();
 await user.click(consent);
 await user.clear(screen.getByLabelText('Requested languages'));
 await user.type(screen.getByLabelText('Requested languages'), 'uk');
 await user.clear(screen.getByLabelText('Primary language'));
 await user.type(screen.getByLabelText('Primary language'), 'uk');
 expect(consent).not.toBeChecked();
 await user.click(consent);
 await user.click(screen.getByRole('button', {name: 'Start analysis'}));
 await waitFor(() => expect(onSubmitted).toHaveBeenCalledTimes(1));
 expect(submitJob).toHaveBeenCalledTimes(1);
 const request = vi.mocked(submitJob).mock.calls[0][0];
 expect(request.configuration).toMatchObject({mode: 'visual', languages: ['uk'], primaryLanguage: 'uk'});
 expect(request.configuration).not.toHaveProperty('context');
 expect(request.consent.configuration).toEqual(request.configuration);
});

it('keeps Context classes unchecked and preserves the exact key while reconciling an ambiguous POST', async () => {
 const user = userEvent.setup();
 vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [aiProfile]});
 vi.mocked(submitJob).mockRejectedValueOnce(new Error('Acknowledgement lost')).mockResolvedValueOnce(aiProgress());
 render(<LaunchForm preview={aiPreview()} onSubmittedAction={() => undefined} />);
 await screen.findByLabelText('Analysis mode');
 await user.selectOptions(screen.getByLabelText('Analysis mode'), 'context-assisted');
 for (const label of ['Capture time', 'Selected album label', 'User hint', 'Nearby known locations (metadata only)']) {
  expect(screen.getByLabelText(label)).not.toBeChecked();
 }
 expect(screen.getByLabelText('Selected album label')).toBeDisabled();
 await user.click(screen.getByLabelText('User hint'));
 await user.type(screen.getByLabelText('Hint'), 'near a bridge');
 await user.click(screen.getByLabelText(/I consent to sending/));
 await user.click(screen.getByRole('button', {name: 'Start analysis'}));
 await screen.findByText(/Acknowledgement lost/);
 await user.click(screen.getByRole('button', {name: 'Reconcile submission'}));
 await screen.findByText(/Analysis accepted/);
 expect(submitJob).toHaveBeenCalledTimes(2);
 const first = vi.mocked(submitJob).mock.calls[0][0];
 expect(vi.mocked(submitJob).mock.calls[1][0]).toEqual(first);
 expect(first.configuration.context).toEqual({version: 'context-v1', classes: ['user_hint'], hint: 'near a bridge'});
 await user.click(screen.getByRole('button', {name: 'Start a different run'}));
 expect(screen.getByLabelText(/I consent to sending/)).not.toBeChecked();
});
