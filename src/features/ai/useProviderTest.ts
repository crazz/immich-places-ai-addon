import {useCallback, useEffect, useRef, useState} from 'react';

import {useCapabilityTestFlight} from './capabilityTestFlight';
import {ProviderAPIError, testProvider} from './providerApi';

import type {TCapabilityReport, TProviderProfile} from './providerApi';

type TProviderTestState = {
	isPending: boolean;
	error: string | null;
	errorCode: string | null;
	report: TCapabilityReport | null;
};

const idle: TProviderTestState = {isPending: false, error: null, errorCode: null, report: null};

export function useProviderTest(profile: TProviderProfile | null): TProviderTestState & {
	start: () => void;
	cancel: () => void;
	clear: () => void;
	isFlightBusy: boolean;
} {
	const flight = useCapabilityTestFlight();
	const acquireFlight = useRef(flight.acquireFlight);
	const releaseFlight = useRef(flight.releaseFlight);
	acquireFlight.current = flight.acquireFlight;
	releaseFlight.current = flight.releaseFlight;
	const [state, setState] = useState<TProviderTestState>(idle);
	const generation = useRef(0);
	const controller = useRef<AbortController | null>(null);
	const holdsFlight = useRef(false);
	const profileID = profile?.id ?? null;
	const revision = profile?.revision ?? null;

	const releaseOwnFlight = useCallback(() => {
		if (!holdsFlight.current) {
			return;
		}
		holdsFlight.current = false;
		releaseFlight.current();
	}, []);

	const clear = useCallback(() => {
		generation.current += 1;
		controller.current?.abort();
		controller.current = null;
		releaseOwnFlight();
		setState(idle);
	}, [releaseOwnFlight]);

	const cancel = useCallback(() => {
		generation.current += 1;
		controller.current?.abort();
		controller.current = null;
		releaseOwnFlight();
		setState(current => current.isPending ? idle : current);
	}, [releaseOwnFlight]);

	useEffect(() => () => {
		generation.current += 1;
		controller.current?.abort();
		releaseOwnFlight();
	}, [releaseOwnFlight]);

	useEffect(() => {
		clear();
	}, [profileID, revision, clear]);

	const start = useCallback(() => {
		if (!profile || !profile.enabled || state.isPending || flight.isFlightBusy) {
			return;
		}
		if (!acquireFlight.current()) {
			return;
		}
		holdsFlight.current = true;
		generation.current += 1;
		const requestGeneration = generation.current;
		controller.current?.abort();
		const next = new AbortController();
		controller.current = next;
		setState({isPending: true, error: null, errorCode: null, report: null});
		testProvider(profile, next.signal).then(report => {
			if (requestGeneration !== generation.current || next.signal.aborted) {
				releaseOwnFlight();
				return;
			}
			releaseOwnFlight();
			setState({isPending: false, error: null, errorCode: null, report});
		}).catch(failure => {
			if (requestGeneration !== generation.current || next.signal.aborted) {
				releaseOwnFlight();
				return;
			}
			releaseOwnFlight();
			if (failure instanceof DOMException && failure.name === 'AbortError') {
				setState(idle);
				return;
			}
			const message = failure instanceof Error ? failure.message : 'Provider test failed.';
			const code = failure instanceof ProviderAPIError ? failure.code : 'REQUEST_FAILED';
			setState({isPending: false, error: message, errorCode: code, report: null});
		});
	}, [profile, state.isPending, flight.isFlightBusy, releaseOwnFlight]);

	return {...state, start, cancel, clear, isFlightBusy: flight.isFlightBusy};
}
