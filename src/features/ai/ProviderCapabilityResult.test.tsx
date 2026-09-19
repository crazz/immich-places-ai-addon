import {render, screen} from '@testing-library/react';
import {expect, it, vi} from 'vitest';

import {ProviderCapabilityResult} from './ProviderCapabilityResult';

import type * as TProviderAPI from './providerApi';
import type {RenderResult} from '@testing-library/react';

const profile: TProviderAPI.TProviderProfile = {
	id: 'profile',
	revision: 2,
	name: 'Private',
	baseURL: 'https://provider.example/v1',
	model: 'manual-model',
	enabled: true,
	hasSecret: true
};

const completedReport: TProviderAPI.TCapabilityReport = {
	attemptID: 'attempt-1',
	profileID: profile.id,
	revision: profile.revision,
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

function renderResult(report: TProviderAPI.TCapabilityReport): RenderResult {
	return render(
		<ProviderCapabilityResult
			profile={profile}
			liveReport={report}
			isPending={false}
			error={null}
			errorCode={null}
			onRetryAction={vi.fn()}
			onReloadAction={vi.fn()}
		/>
	);
}

it('reports a current-revision failed result as failed instead of stale', () => {
	renderResult({
		...completedReport,
		lifecycle: 'failed',
		compatibility: 'failed',
		applicable: false,
		observations: {
			image: {status: 'unverified', reason: 'authentication'},
			json: {status: 'unverified'},
			strict: {status: 'unverified'}
		}
	});
	expect(screen.getByText(/Failed: failed/i)).toBeVisible();
	expect(screen.queryByText(/Stale/i)).not.toBeInTheDocument();
});

it('reports a current-revision interrupted result as interrupted instead of stale', () => {
	renderResult({
		...completedReport,
		lifecycle: 'interrupted',
		compatibility: 'incomplete',
		applicable: false,
		observations: {
			image: {status: 'unverified', reason: 'interrupted'},
			json: {status: 'unverified'},
			strict: {status: 'unverified'}
		}
	});
	expect(screen.getByText(/Interrupted: reload the report/i)).toBeVisible();
	expect(screen.queryByText(/Stale/i)).not.toBeInTheDocument();
});

it('reports a current-revision unsupported result as unsupported instead of stale', () => {
	renderResult({
		...completedReport,
		compatibility: 'unsupported',
		applicable: false,
		observations: {
			image: {status: 'unsupported', reason: 'unsupported_mode'},
			json: {status: 'unverified'},
			strict: {status: 'unverified'}
		}
	});
	expect(screen.getByText(/Unsupported for this protocol/i)).toBeVisible();
	expect(screen.queryByText(/Stale/i)).not.toBeInTheDocument();
});

it('reports a completed result that no longer applies to the current configuration as stale', () => {
	renderResult({...completedReport, applicable: false});
	expect(screen.getByText(/Stale: result is not current for this profile revision or policy/i)).toBeVisible();
	expect(screen.getByRole('button', {name: /Retry test/i})).toBeVisible();
});

it('reports a result from an earlier profile revision as stale', () => {
	renderResult({...completedReport, revision: profile.revision - 1});
	expect(screen.getByText(/Stale: result is not current for this profile revision or policy/i)).toBeVisible();
	expect(screen.getByText(/Tested revision 1/)).toBeVisible();
});

it('explains a canceled result separately from an interrupted one', () => {
	renderResult({
		...completedReport,
		lifecycle: 'canceled',
		compatibility: 'failed',
		applicable: false,
		observations: {
			image: {status: 'unverified', reason: 'canceled'},
			json: {status: 'unverified'},
			strict: {status: 'unverified'}
		}
	});
	expect(screen.getByText(/Canceled: the test stopped before completing/i)).toBeVisible();
	expect(screen.queryByText(/Interrupted/i)).not.toBeInTheDocument();
	expect(screen.getByRole('button', {name: /Retry test/i})).toBeVisible();
});

it('shows the observed compatibility summary of a completed result', () => {
	renderResult({
		...completedReport,
		compatibility: 'json-only compatible',
		observations: {
			image: {status: 'supported'},
			json: {status: 'supported'},
			strict: {status: 'unsupported'}
		}
	});
	expect(screen.getByText(/Compatibility: json-only compatible/i)).toBeVisible();
});
