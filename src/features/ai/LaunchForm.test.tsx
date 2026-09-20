import {render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import {submitJob} from './jobApi';
import {LaunchForm} from './LaunchForm';
import {fetchProviders} from './providerApi';
import {aiPreview, aiProfile, aiProgress} from './testing/fixtures';

import type {TProviderProfile} from './providerApi';

vi.mock('./providerApi', () => ({fetchProviders: vi.fn()}));
vi.mock('./jobApi', () => ({submitJob: vi.fn()}));
afterEach(() => { vi.resetAllMocks(); vi.unstubAllGlobals(); });

it('starts and reconciles a run when the browser has no randomUUID API', async () => {
 const user = userEvent.setup();
 vi.stubGlobal('crypto', {getRandomValues: crypto.getRandomValues.bind(crypto)});
 vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [aiProfile]});
 vi.mocked(submitJob).mockRejectedValueOnce(new Error('Acknowledgement lost')).mockResolvedValue(aiProgress());
 render(<LaunchForm preview={aiPreview()} onSubmittedAction={() => undefined} />);
 await user.click(await screen.findByRole('button', {name: 'Start analysis'}));
 await waitFor(() => expect(submitJob).toHaveBeenCalledTimes(1));
 const first = vi.mocked(submitJob).mock.calls[0][0];
 expect(first.idempotencyKey).toMatch(/^[a-f0-9]{32}$/);
 await user.click(screen.getByRole('button', {name: 'Reconcile submission'}));
 await screen.findByText(/Analysis accepted/);
 expect(vi.mocked(submitJob).mock.calls[1][0]).toEqual(first);
 await user.click(screen.getByRole('button', {name: 'Start a different run'}));
 await user.click(screen.getByRole('button', {name: 'Start analysis'}));
 await waitFor(() => expect(submitJob).toHaveBeenCalledTimes(3));
 expect(vi.mocked(submitJob).mock.calls[2][0].idempotencyKey).not.toBe(first.idempotencyKey);
});

it('reports a local preparation failure without claiming a run exists and allows retry', async () => {
 const user = userEvent.setup();
 const getRandomValues = crypto.getRandomValues.bind(crypto);
 vi.stubGlobal('crypto', {getRandomValues: () => { throw new Error('Random source unavailable'); }});
 vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [aiProfile]});
 vi.mocked(submitJob).mockResolvedValue(aiProgress());
 render(<LaunchForm preview={aiPreview()} onSubmittedAction={() => undefined} />);
 await user.click(await screen.findByRole('button', {name: 'Start analysis'}));
 expect(screen.getByRole('alert')).toHaveTextContent(/^Random source unavailable$/);
 expect(screen.queryByRole('button', {name: 'Reconcile submission'})).not.toBeInTheDocument();
 expect(submitJob).not.toHaveBeenCalled();
 vi.stubGlobal('crypto', {getRandomValues});
 await user.click(screen.getByRole('button', {name: 'Start analysis'}));
 await screen.findByText(/Analysis accepted/);
 expect(submitJob).toHaveBeenCalledTimes(1);
 expect(screen.queryByRole('alert')).not.toBeInTheDocument();
});

it('starts the displayed Visual run with one click and binds the latest choices', async () => {
 const user = userEvent.setup();
 vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [aiProfile]});
 vi.mocked(submitJob).mockResolvedValue(aiProgress());
 const onSubmitted = vi.fn();
 render(<LaunchForm preview={aiPreview()} onSubmittedAction={onSubmitted} />);
 await screen.findByLabelText('Requested languages');
 await user.selectOptions(screen.getByLabelText('Analysis mode'), 'visual');
 expect(screen.queryByLabelText(/I consent to sending/)).not.toBeInTheDocument();
 expect(screen.getByRole('button', {name: 'Start analysis'})).toBeEnabled();
 expect(submitJob).not.toHaveBeenCalled();
 await user.clear(screen.getByLabelText('Requested languages'));
 await user.type(screen.getByLabelText('Requested languages'), 'uk');
 await user.clear(screen.getByLabelText('Primary language'));
 await user.type(screen.getByLabelText('Primary language'), 'uk');
 await user.click(screen.getByRole('button', {name: 'Start analysis'}));
 await waitFor(() => expect(onSubmitted).toHaveBeenCalledTimes(1));
 expect(submitJob).toHaveBeenCalledTimes(1);
 const request = vi.mocked(submitJob).mock.calls[0][0];
 expect(request.configuration).toMatchObject({mode: 'visual', languages: ['uk'], primaryLanguage: 'uk'});
 expect(request.configuration).not.toHaveProperty('context');
 expect(request.consent.configuration).toEqual(request.configuration);
 expect(request.consent.image).toBe(true);
});

it('keeps advanced controls optional and refreshes working defaults when the provider changes', async () => {
 const user = userEvent.setup();
 const jsonProvider: TProviderProfile = {...aiProfile, id: 'json-provider', name: 'JSON provider', executionReadiness: {...aiProfile.executionReadiness!, source: 'application-defaults', maxInputTokens: 200000, maxOutputTokens: 8000}, capabilityReport: {...aiProfile.capabilityReport!, observations: {...aiProfile.capabilityReport!.observations, strict: {status: 'unsupported'}}}};
 vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [aiProfile, jsonProvider]});
 vi.mocked(submitJob).mockResolvedValue(aiProgress());
 render(<LaunchForm preview={aiPreview()} onSubmittedAction={() => undefined} />);
 await screen.findByLabelText('Provider');
 expect(screen.getByLabelText('Output tokens per call')).not.toBeVisible();
 await user.selectOptions(screen.getByLabelText('Provider'), jsonProvider.id);
 expect(screen.getByRole('button', {name: 'Start analysis'})).toBeEnabled();
 expect(submitJob).not.toHaveBeenCalled();
 await user.click(screen.getByText('Advanced settings'));
 expect(screen.getByLabelText('Output tokens per call')).toHaveValue(4000);
 expect(screen.getByLabelText('Total token allowance')).toHaveValue(204000 * aiPreview().eligibleCount);
 await user.click(screen.getByRole('button', {name: 'Start analysis'}));
 await waitFor(() => expect(submitJob).toHaveBeenCalledTimes(1));
 expect(vi.mocked(submitJob).mock.calls[0][0].configuration).toMatchObject({profileId: jsonProvider.id, format: 'json', allowJson: true, limits: {outputTokens: 4000}});
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
 await user.click(screen.getByRole('button', {name: 'Start analysis'}));
 await screen.findByText(/Acknowledgement lost/);
 await user.click(screen.getByRole('button', {name: 'Reconcile submission'}));
 await screen.findByText(/Analysis accepted/);
 expect(submitJob).toHaveBeenCalledTimes(2);
 const first = vi.mocked(submitJob).mock.calls[0][0];
 expect(vi.mocked(submitJob).mock.calls[1][0]).toEqual(first);
 expect(first.configuration.context).toEqual({version: 'context-v1', classes: ['user_hint'], hint: 'near a bridge'});
 await user.click(screen.getByRole('button', {name: 'Start a different run'}));
 expect(screen.getByRole('button', {name: 'Start analysis'})).toBeEnabled();
 expect(submitJob).toHaveBeenCalledTimes(2);
});
