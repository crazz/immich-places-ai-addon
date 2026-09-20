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

it('starts Research by default with the exact hint and no extra confirmation or hidden context', async () => {
	const user = userEvent.setup();
	vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [aiProfile]});
	vi.mocked(submitJob).mockResolvedValue(aiProgress());
	render(
		<LaunchForm
			preview={aiPreview()}
			onSubmittedAction={() => undefined}
		/>
	);
	expect(await screen.findByLabelText('Analysis mode')).toHaveValue('research');
	expect(screen.queryByLabelText('Nearby known locations (metadata only)')).not.toBeInTheDocument();
	expect(screen.queryByLabelText('User hint')).not.toBeInTheDocument();
	await user.type(screen.getByLabelText('Hint'), 'Maybe Acre? Compare the church facade.');
	expect(submitJob).not.toHaveBeenCalled();
	await user.click(screen.getByRole('button', {name: 'Start analysis'}));
	await waitFor(() => expect(submitJob).toHaveBeenCalledTimes(1));
	const request = vi.mocked(submitJob).mock.calls[0][0];
	expect(request.configuration).toMatchObject({
		mode: 'research',
		context: {version: 'context-v1', classes: ['user_hint'], hint: 'Maybe Acre? Compare the church facade.'},
		limits: {maxCalls: 1}
	});
	expect(request.consent.configuration).toEqual(request.configuration);
});

it('shows optional album and recorded capture values and includes only selected classes', async () => {
	const user = userEvent.setup();
	vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [aiProfile]});
	vi.mocked(submitJob).mockResolvedValue(aiProgress());
	const preview = aiPreview();
	preview.scope = {...preview.scope, view: 'album', albumID: '00000000-0000-4000-8000-000000000080'};
	const contextPreview = {albumLabel: 'Israel trip', captureTimes: {[preview.assetIDs[0]]: '2009-04-01T12:30:00'}};
	render(
		<LaunchForm
			preview={{...preview, contextPreview}}
			onSubmittedAction={() => undefined}
		/>
	);
	await screen.findByLabelText('Analysis mode');
	expect(screen.getByText('Israel trip')).toBeVisible();
	expect(screen.getByText('2009-04-01T12:30:00')).toBeVisible();
	await user.click(screen.getByRole('checkbox', {name: 'Include selected album label'}));
	await user.click(screen.getByRole('checkbox', {name: 'Include recorded capture times'}));
	await user.click(screen.getByRole('button', {name: 'Start analysis'}));
	await waitFor(() => expect(submitJob).toHaveBeenCalledTimes(1));
	expect(vi.mocked(submitJob).mock.calls[0][0].configuration.context).toEqual({
		version: 'context-v1',
		classes: ['selected_album', 'capture_time'],
		albumId: preview.scope.albumID
	});
});
