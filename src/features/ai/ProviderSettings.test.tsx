import {act, fireEvent, render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import {ProviderAPIError, fetchProviders, saveProvider, testProvider} from './providerApi';
import {ProviderSettings} from './ProviderSettings';

import type * as TProviderAPI from './providerApi';

vi.mock('./providerApi', async importOriginal => ({
	...await importOriginal<typeof TProviderAPI>(),
	fetchProviders: vi.fn(),
	saveProvider: vi.fn(),
	testProvider: vi.fn()
}));
afterEach(() => vi.resetAllMocks());

const profile = {id: 'profile', revision: 1, name: 'Private', baseURL: 'https://provider.example/v1', model: 'manual-model', enabled: true, hasSecret: true};

const completedReport: TProviderAPI.TCapabilityReport = {
	attemptID: 'attempt-1',
	profileID: profile.id,
	revision: 1,
	protocolVersion: 'capability-v1',
	policyFingerprint: 'fp',
	lifecycle: 'completed',
	startedAt: '2026-09-19T12:00:00.000Z',
	deadlineAt: '2026-09-19T12:02:00.000Z',
	completedAt: '2026-09-19T12:00:05.000Z',
	requestedModel: profile.model,
	observations: {
		image: {status: 'supported'},
		json: {status: 'supported'},
		strict: {status: 'supported'}
	},
	compatibility: 'strict-schema sample compatible',
	applicable: true
};

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

it('starts and inspects a test using the keyboard', async () => {
	const user = userEvent.setup();
	vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [profile]});
	const pending = Promise.withResolvers<TProviderAPI.TCapabilityReport>();
	vi.mocked(testProvider).mockReturnValue(pending.promise);
	render(<ProviderSettings />);
	await user.click(screen.getByRole('button', {name: 'AI providers'}));
	expect(await screen.findByText(profile.baseURL)).toBeVisible();
	expect(screen.getByText(new RegExp(profile.model))).toBeVisible();
	expect(screen.getByText(/Synthetic-image test of this destination\/model/i)).toBeVisible();
	expect(screen.getByText(/provider usage/i)).toBeVisible();
	expect(screen.getByText(/local proxy can forward .* to its upstream model service/i)).toBeVisible();
	expect(screen.queryByText(/geolocation quality/i)).not.toBeInTheDocument();
	expect(screen.queryByText(/private photograph/i)).not.toBeInTheDocument();
	const testAction = await screen.findByRole('button', {name: 'Test provider'});
	testAction.focus();
	expect(testAction).toHaveFocus();
	await user.keyboard('{Enter}');
	expect(await screen.findByRole('status')).toHaveTextContent(/Testing provider/i);
	expect(screen.getByRole('button', {name: 'Test provider'})).toBeDisabled();
	expect(testProvider).toHaveBeenCalledExactlyOnceWith(profile, expect.any(AbortSignal));
	const persisted = {...profile, capabilityReport: completedReport};
	vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [persisted]});
	await act(async () => pending.resolve(completedReport));
	expect(await screen.findByText(/Image: supported/i)).toBeVisible();
	expect(screen.getByText(/JSON: supported/i)).toBeVisible();
	expect(screen.getByText(/Strict: supported/i)).toBeVisible();
	expect(screen.getByText(/Revision 1/)).toBeVisible();
	await user.click(screen.getByRole('button', {name: 'Reload profiles'}));
	await screen.findByText(/Image: supported/i);
	expect(testProvider).toHaveBeenCalledTimes(1);
});

it('handles cancel and stale responses safely', async () => {
	const user = userEvent.setup();
	vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [profile]});
	const first = Promise.withResolvers<TProviderAPI.TCapabilityReport>();
	let firstSignal: AbortSignal | undefined;
	vi.mocked(testProvider).mockImplementation(async (_profile, signal) => {
		firstSignal = signal;
		return first.promise;
	});
	render(<ProviderSettings />);
	await user.click(screen.getByRole('button', {name: 'AI providers'}));
	await user.click(await screen.findByRole('button', {name: 'Test provider'}));
	expect(await screen.findByRole('status')).toHaveTextContent(/Testing provider/i);
	expect(screen.getByRole('button', {name: 'Test provider'})).toBeDisabled();
	await user.click(screen.getByRole('button', {name: 'Test provider'}));
	expect(testProvider).toHaveBeenCalledTimes(1);
	await user.click(screen.getByRole('button', {name: 'Close dialog'}));
	expect(firstSignal?.aborted).toBe(true);
	await act(async () => first.resolve(completedReport));

	const second = Promise.withResolvers<TProviderAPI.TCapabilityReport>();
	let secondSignal: AbortSignal | undefined;
	vi.mocked(testProvider).mockImplementation(async (_profile, signal) => {
		secondSignal = signal;
		return second.promise;
	});
	await user.click(screen.getByRole('button', {name: 'AI providers'}));
	await user.click(await screen.findByRole('button', {name: 'Test provider'}));
	expect(await screen.findByRole('status')).toHaveTextContent(/Testing provider/i);
	await user.click(screen.getByRole('button', {name: 'Edit Private'}));
	expect(secondSignal?.aborted).toBe(true);
	await act(async () => second.resolve(completedReport));
	await user.click(screen.getByRole('button', {name: 'Cancel'}));
	expect(screen.queryByText(/Image: supported/i)).not.toBeInTheDocument();
	expect(screen.getByRole('button', {name: 'Test provider'})).toBeEnabled();
});

it('shows a returned failed report without reloading profiles', async () => {
	const user = userEvent.setup();
	vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [profile]});
	vi.mocked(testProvider).mockResolvedValue({
		...completedReport,
		lifecycle: 'failed',
		compatibility: 'failed',
		applicable: false,
		observations: {
			image: {status: 'supported'},
			json: {status: 'unverified', reason: 'policy'},
			strict: {status: 'unverified'}
		}
	});
	render(<ProviderSettings />);
	await user.click(screen.getByRole('button', {name: 'AI providers'}));
	await user.click(await screen.findByRole('button', {name: 'Test provider'}));
	expect(await screen.findByText(/Failed: failed/i)).toBeVisible();
	expect(screen.getByText(/Image: supported/i)).toBeVisible();
	expect(screen.getByText(/JSON: unverified/i)).toBeVisible();
	expect(screen.getByRole('button', {name: /Retry test/i})).toBeVisible();
	expect(fetchProviders).toHaveBeenCalledTimes(1);
});

it('explains failures without exposing secrets or inventing readiness', async () => {
	const user = userEvent.setup();
	const secret = 'never-leak-secret';
	const cases: {code: string; message: string; action: RegExp; report?: TProviderAPI.TCapabilityReport}[] = [
		{code: 'PROVIDER_BUSY', message: 'another capability test is already running', action: /Reload report/i},
		{code: 'PROVIDER_FAILURE', message: 'capability test could not contact the provider', action: /Retry test/i},
		{
			code: '',
			message: '',
			action: /Reload report/i,
			report: {
				...completedReport,
				lifecycle: 'interrupted',
				applicable: false,
				compatibility: 'incomplete',
				observations: {
					image: {status: 'unverified', reason: 'interrupted'},
					json: {status: 'unverified'},
					strict: {status: 'unverified'}
				}
			}
		},
		{
			code: '',
			message: '',
			action: /Retry test/i,
			report: {
				...completedReport,
				lifecycle: 'completed',
				applicable: false,
				compatibility: 'unsupported',
				observations: {
					image: {status: 'unsupported', reason: 'unsupported_mode'},
					json: {status: 'unverified'},
					strict: {status: 'unverified'}
				}
			}
		},
		{
			code: '',
			message: '',
			action: /Retry test/i,
			report: {
				...completedReport,
				lifecycle: 'failed',
				applicable: false,
				compatibility: 'failed',
				observations: {
					image: {status: 'unverified', reason: 'authentication'},
					json: {status: 'unverified'},
					strict: {status: 'unverified'}
				}
			}
		},
		{
			code: '',
			message: '',
			action: /Retry test/i,
			report: {
				...completedReport,
				lifecycle: 'completed',
				applicable: false,
				compatibility: 'incomplete',
				observations: {
					image: {status: 'supported'},
					json: {status: 'unverified'},
					strict: {status: 'unverified'}
				}
			}
		}
	];
	for (const entry of cases) {
		vi.mocked(fetchProviders).mockResolvedValue({
			enabled: true,
			items: [entry.report ? {...profile, capabilityReport: entry.report} : profile]
		});
		if (entry.code) {
			vi.mocked(testProvider).mockRejectedValue(new ProviderAPIError(entry.message, entry.code));
		}
		const view = render(<ProviderSettings />);
		await user.click(screen.getByRole('button', {name: 'AI providers'}));
		if (entry.code) {
			await user.click(await screen.findByRole('button', {name: 'Test provider'}));
			expect(await screen.findByRole('alert')).toHaveTextContent(entry.message);
			expect(screen.getByRole('button', {name: entry.action})).toBeVisible();
		} else {
			expect(await screen.findByText(/Image:/i)).toBeVisible();
			expect(screen.getByText(/Observed synthetic compatibility only/i)).toBeVisible();
			expect(screen.getByText(/not geolocation quality or private-photo authorization/i)).toBeVisible();
			expect(screen.getByRole('button', {name: entry.action})).toBeVisible();
		}
		expect(screen.queryByText(secret)).not.toBeInTheDocument();
		expect(screen.queryByText(/choices/i)).not.toBeInTheDocument();
		expect(screen.queryByText(/Bearer/i)).not.toBeInTheDocument();
		view.unmount();
	}
});

it('aborts pending capability tests for every profile when the dialog closes', async () => {
	const user = userEvent.setup();
	const other = {...profile, id: 'other', name: 'Other', model: 'other-model'};
	vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [profile, other]});
	const pending = Promise.withResolvers<TProviderAPI.TCapabilityReport>();
	let pendingSignal: AbortSignal | undefined;
	vi.mocked(testProvider).mockImplementation(async (_profile, signal) => {
		pendingSignal = signal;
		return pending.promise;
	});
	render(<ProviderSettings />);
	await user.click(screen.getByRole('button', {name: 'AI providers'}));
	const testButtons = await screen.findAllByRole('button', {name: 'Test provider'});
	expect(testButtons).toHaveLength(2);
	await user.click(testButtons[0]);
	expect(await screen.findByRole('status')).toHaveTextContent(/Testing provider/i);
	expect(testButtons[1]).toBeDisabled();
	await user.click(screen.getByRole('button', {name: 'Close dialog'}));
	expect(pendingSignal?.aborted).toBe(true);
});

it('blocks starting a second profile test while another is pending', async () => {
	const user = userEvent.setup();
	const other = {...profile, id: 'other', name: 'Other', model: 'other-model'};
	vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [profile, other]});
	const pending = Promise.withResolvers<TProviderAPI.TCapabilityReport>();
	vi.mocked(testProvider).mockReturnValue(pending.promise);
	render(<ProviderSettings />);
	await user.click(screen.getByRole('button', {name: 'AI providers'}));
	const testButtons = await screen.findAllByRole('button', {name: 'Test provider'});
	expect(testButtons).toHaveLength(2);
	testButtons[0].focus();
	await user.keyboard('{Enter}');
	expect(await screen.findByRole('status')).toHaveTextContent(/Testing provider/i);
	expect(testProvider).toHaveBeenCalledTimes(1);
	expect(testButtons[0]).toBeDisabled();
	expect(testButtons[1]).toBeDisabled();
	await user.click(testButtons[1]);
	expect(testProvider).toHaveBeenCalledTimes(1);
	await act(async () => pending.resolve(completedReport));
	await waitFor(() => {
		const buttons = screen.getAllByRole('button', {name: 'Test provider'});
		expect(buttons).toHaveLength(2);
		expect(buttons.every(button => !(button as HTMLButtonElement).disabled)).toBe(true);
	});
});
