import {act, fireEvent, render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import {ProviderAPIError, fetchProviders, saveProvider} from './providerApi';
import {ProviderSettings} from './ProviderSettings';

import type * as TProviderAPI from './providerApi';

vi.mock('./providerApi', async importOriginal => ({...await importOriginal<typeof TProviderAPI>(), fetchProviders: vi.fn(), saveProvider: vi.fn()}));
afterEach(() => vi.resetAllMocks());

const profile = {id: 'profile', revision: 1, name: 'Private', baseURL: 'https://provider.example/v1', model: 'manual-model', enabled: true, hasSecret: true};

it('creates, edits and disables a private profile with labeled controls and redacted secrets', async () => {
	const user = userEvent.setup();
	vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: []});
	render(<ProviderSettings />);
	await user.click(screen.getByRole('button', {name: 'AI providers'}));
	await user.click(await screen.findByRole('button', {name: 'Create provider'}));
	await user.type(screen.getByLabelText('Name'), 'Private');
	await user.type(screen.getByLabelText('API base URL'), profile.baseURL);
	await user.type(screen.getByLabelText('Model'), profile.model);
	await user.type(screen.getByLabelText('API key'), 'synthetic-secret');
	expect(screen.getByLabelText('API key')).toHaveAttribute('type', 'password');
	vi.mocked(saveProvider).mockResolvedValue(profile);
	vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [profile]});
	await user.click(screen.getByRole('button', {name: 'Save provider'}));
	await screen.findByText('Saved revision 1');
	expect(saveProvider).toHaveBeenCalledWith({name: 'Private', baseURL: profile.baseURL, model: profile.model, enabled: true, secret: 'synthetic-secret'}, undefined);
	await user.click(await screen.findByRole('button', {name: 'Edit Private'}));
	expect(screen.getByLabelText('API key')).toHaveValue('');
	await user.click(screen.getByLabelText('Enabled'));
	const disabled = {...profile, enabled: false, revision: 2};
	vi.mocked(saveProvider).mockResolvedValue(disabled);
	vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [disabled]});
	await user.click(screen.getByRole('button', {name: 'Save provider'}));
	await waitFor(() => expect(saveProvider).toHaveBeenLastCalledWith({name: 'Private', baseURL: profile.baseURL, model: profile.model, enabled: false, secret: undefined}, profile));
	await screen.findByText('Saved revision 2');
	expect(await screen.findByText(/Disabled · Revision 2/)).toBeVisible();
});

it('discards cancelled keys and blocks duplicate saves while preserving safe edits after a conflict', async () => {
	const user = userEvent.setup();
	vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [profile]});
	render(<ProviderSettings />);
	await user.click(screen.getByRole('button', {name: 'AI providers'}));
	await user.click(await screen.findByRole('button', {name: 'Edit Private'}));
	await user.type(screen.getByLabelText('API key'), 'cancelled-key');
	await user.click(screen.getByRole('button', {name: 'Cancel'}));
	await user.click(screen.getByRole('button', {name: 'Edit Private'}));
	expect(screen.getByLabelText('API key')).toHaveValue('');
	await user.type(screen.getByLabelText('API key'), 'closed-key');
	await user.click(screen.getByRole('button', {name: 'Close dialog'}));
	await user.click(screen.getByRole('button', {name: 'AI providers'}));
	await user.click(await screen.findByRole('button', {name: 'Edit Private'}));
	expect(screen.getByLabelText('API key')).toHaveValue('');
	expect(saveProvider).not.toHaveBeenCalled();
	await user.clear(screen.getByLabelText('Name'));
	await user.type(screen.getByLabelText('Name'), 'Unsaved name');
	await user.type(screen.getByLabelText('API key'), 'submitted-key');
	const pending = Promise.withResolvers<never>();
	vi.mocked(saveProvider).mockReturnValue(pending.promise);
	await user.click(screen.getByRole('button', {name: 'Save provider'}));
	expect(screen.getByRole('button', {name: 'Saving…'})).toBeDisabled();
	expect(screen.getByRole('button', {name: 'Cancel'})).toBeDisabled();
	expect(screen.getByRole('button', {name: 'Reload profiles'})).toBeDisabled();
	fireEvent.submit(screen.getByRole('form', {name: 'Provider profile'}));
	await user.click(screen.getByRole('button', {name: 'Close dialog'}));
	expect(screen.getByRole('dialog')).toBeVisible();
	expect(saveProvider).toHaveBeenCalledTimes(1);
	await act(async () => pending.reject(new ProviderAPIError('Provider changed; reload before editing', 'PROVIDER_CONFLICT')));
	expect(await screen.findByRole('alert')).toHaveTextContent('Provider changed; reload before editing');
	expect(screen.getByLabelText('Name')).toHaveValue('Unsaved name');
	expect(screen.getByLabelText('API key')).toHaveValue('');
	expect(saveProvider).toHaveBeenCalledTimes(1);
	vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [{...profile, name: 'Updated elsewhere', revision: 2}]});
	await user.click(screen.getByRole('button', {name: 'Reload profiles'}));
	await screen.findByRole('button', {name: 'Edit Updated elsewhere'});
	expect(saveProvider).toHaveBeenCalledTimes(1);
});

it('requires explicit removal and sends no retained key to a changed destination', async () => {
	const user = userEvent.setup();
	vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [profile]});
	render(<ProviderSettings />);
	await user.click(screen.getByRole('button', {name: 'AI providers'}));
	await user.click(await screen.findByRole('button', {name: 'Edit Private'}));
	await user.clear(screen.getByLabelText('API base URL'));
	await user.type(screen.getByLabelText('API base URL'), 'https://replacement.example/v1');
	expect(screen.getByText('Changing the URL requires a replacement key or explicit removal.')).toBeVisible();
	await user.type(screen.getByLabelText('API key'), 'discarded-replacement');
	await user.click(screen.getByLabelText('Remove all stored keys for this profile'));
	expect(screen.getByLabelText('API key')).toBeDisabled();
	expect(screen.getByLabelText('API key')).toHaveValue('');
	vi.mocked(saveProvider).mockRejectedValue(new Error('Temporary storage failure'));
	await user.click(screen.getByRole('button', {name: 'Save provider'}));
	expect(await screen.findByRole('alert')).toHaveTextContent('Temporary storage failure');
	expect(saveProvider).toHaveBeenCalledExactlyOnceWith({name: profile.name, model: profile.model, enabled: true, secret: '', baseURL: 'https://replacement.example/v1'}, profile);
});

it('shows loading, load failure, disabled and empty states with explicit reload', async () => {
	const user = userEvent.setup();
	const pending = Promise.withResolvers<never>();
	vi.mocked(fetchProviders).mockReturnValue(pending.promise);
	render(<ProviderSettings />);
	await user.click(screen.getByRole('button', {name: 'AI providers'}));
	expect(screen.getByRole('status')).toHaveTextContent('Loading provider settings…');
	await act(async () => pending.reject(new Error('offline')));
	expect(await screen.findByRole('alert')).toHaveTextContent('Could not load provider settings.');
	vi.mocked(fetchProviders).mockResolvedValue({enabled: false, items: []});
	await user.click(screen.getByRole('button', {name: 'Reload profiles'}));
	await screen.findByText('AI is disabled by the administrator.');
	expect(screen.queryByRole('button', {name: 'Create provider'})).not.toBeInTheDocument();
	vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: []});
	await user.click(screen.getByRole('button', {name: 'Reload profiles'}));
	await screen.findByText('No private providers yet.');
	expect(screen.getByRole('button', {name: 'Create provider'})).toBeEnabled();
	expect(saveProvider).not.toHaveBeenCalled();
});
