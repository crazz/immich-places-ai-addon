import {useEffect, useState} from 'react';

import {fetchProviders} from './providerApi';

import type {TProviderList} from './providerApi';

type TProviderListState = {data: TProviderList | null; error: string | null; isLoading: boolean};

export function useProviderList(): TProviderListState & {reload: () => void} {
	const [state, setState] = useState<TProviderListState>({data: null, error: null, isLoading: true});
	const [generation, setGeneration] = useState(0);
	useEffect(() => {
		const controller = new AbortController();
		setState({data: null, error: null, isLoading: true});
		fetchProviders(controller.signal).then(data => {
			if (!controller.signal.aborted) {
				setState({data, error: null, isLoading: false});
			}
		}).catch(() => {
			if (!controller.signal.aborted) {
				setState({data: null, error: 'Could not load provider settings.', isLoading: false});
			}
		});
		return () => controller.abort();
	}, [generation]);
	return {...state, reload: () => setGeneration(value => value + 1)};
}
