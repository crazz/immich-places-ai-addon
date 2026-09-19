'use client';

import {createContext, useCallback, useContext, useMemo, useRef, useState} from 'react';

import type {ReactElement, ReactNode} from 'react';

type TCapabilityTestFlight = {
	isFlightBusy: boolean;
	acquireFlight: () => boolean;
	releaseFlight: () => void;
};

const CapabilityTestFlightContext = createContext<TCapabilityTestFlight | null>(null);

const idleFlight: TCapabilityTestFlight = {
	isFlightBusy: false,
	acquireFlight: () => true,
	releaseFlight: () => undefined
};

export function CapabilityTestFlightProvider({children}: {children: ReactNode}): ReactElement {
	const [isFlightBusy, setIsFlightBusy] = useState(false);
	const isBusyRef = useRef(false);
	const acquireFlight = useCallback(() => {
		if (isBusyRef.current) {
			return false;
		}
		isBusyRef.current = true;
		setIsFlightBusy(true);
		return true;
	}, []);
	const releaseFlight = useCallback(() => {
		isBusyRef.current = false;
		setIsFlightBusy(false);
	}, []);
	const value = useMemo(() => ({isFlightBusy, acquireFlight, releaseFlight}), [isFlightBusy, acquireFlight, releaseFlight]);
	return <CapabilityTestFlightContext.Provider value={value}>{children}</CapabilityTestFlightContext.Provider>;
}

export function useCapabilityTestFlight(): TCapabilityTestFlight {
	return useContext(CapabilityTestFlightContext) ?? idleFlight;
}
