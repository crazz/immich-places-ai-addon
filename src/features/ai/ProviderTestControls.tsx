import {useEffect} from 'react';

import {ProviderCapabilityResult} from './ProviderCapabilityResult';
import {useProviderTest} from './useProviderTest';

import type {TProviderProfile} from './providerApi';
import type {ReactElement} from 'react';

export function ProviderTestControls({
	profile,
	onReloadAction,
	onRegisterCancelAction
}: {
	profile: TProviderProfile;
	onReloadAction: () => void;
	onRegisterCancelAction?: (cancel: () => void) => () => void;
}): ReactElement {
	const test = useProviderTest(profile.enabled ? profile : null);
	useEffect(() => {
		if (!onRegisterCancelAction) {
			return;
		}
		return onRegisterCancelAction(test.cancel);
	}, [onRegisterCancelAction, test.cancel]);
	return (
		<div className={'mt-2 space-y-2'}>
			<p className={'text-xs text-(--color-text-secondary)'}>
				{'Synthetic-image test of this destination/model. Up to three requests may consume provider usage. The local proxy can forward this synthetic input to its upstream model service. Cancellation stops waiting locally and cannot reverse upstream usage.'}
			</p>
			{profile.enabled && (
				<button
					type={'button'}
					className={'text-sm underline disabled:opacity-50'}
					disabled={test.isFlightBusy}
					onClick={test.start}>
					{'Test provider'}
				</button>
			)}
			<ProviderCapabilityResult
				profile={profile}
				liveReport={test.report}
				isPending={test.isPending}
				error={test.error}
				errorCode={test.errorCode}
				onRetryAction={test.start}
				onReloadAction={onReloadAction}
			/>
		</div>
	);
}
