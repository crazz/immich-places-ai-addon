import {useEffect, useState} from 'react';

import {AIRequestError} from './aiRequest';
import {fetchJob} from './jobApi';

import type {TJobProgress} from './jobTypes';

type TProgressState = {key: string; data: TJobProgress | null; stale: boolean; error: string; loading: boolean};

export function jobIsActive(job: TJobProgress): boolean {
	return !job.blocked && ['queued', 'running', 'retry_wait'].some(state => (job.counts[state] ?? 0) > 0);
}

export function useJobProgress(owner: string, id: string | null): Omit<TProgressState, 'key'> & {refresh: () => void} {
	const key = `${owner}:${id ?? ''}`;
	const [state, setState] = useState<TProgressState>({key, data: null, stale: false, error: '', loading: !!id});
	const [generation, setGeneration] = useState(0);
	useEffect(() => {
		let isAlive = true;
		let flight: AbortController | null = null;
		let timer: ReturnType<typeof setTimeout> | undefined;
		let failures = 0;
		let isTerminal = false;
		setState({key, data: null, stale: false, error: '', loading: !!id});
		const stopFlight = (): void => { clearTimeout(timer); flight?.abort(); };
		const poll = async (): Promise<void> => {
			stopFlight();
			if (!isAlive || !owner || !id || isTerminal) { return; }
			if (document.hidden || !navigator.onLine) {
				setState(current => ({...current, stale: true, loading: false, error: navigator.onLine ? 'Updates paused while hidden.' : 'Offline. Progress may be stale.'}));
				return;
			}
			const controller = new AbortController(); flight = controller;
			try {
				const data = await fetchJob(id, controller.signal);
				if (!isAlive || controller.signal.aborted) { return; }
				failures = 0;
				isTerminal = !jobIsActive(data);
				setState({key, data, stale: false, error: '', loading: false});
				if (!isTerminal) { timer = setTimeout(() => void poll(), 2000); }
			} catch (error) {
				if (!isAlive || controller.signal.aborted) { return; }
				setState(current => ({...current, stale: true, loading: false, error: 'Could not refresh progress. Displayed progress may be stale.'}));
				if (error instanceof AIRequestError && ['UNAUTHENTICATED', 'JOB_UNAVAILABLE'].includes(error.code)) { return; }
				failures++;
				timer = setTimeout(() => void poll(), Math.min(30000, 2000 * 2 ** Math.min(failures, 4)));
			}
		};
		const visibility = (): void => { void poll(); };
		void poll();
		document.addEventListener('visibilitychange', visibility);
		window.addEventListener('online', visibility);
		window.addEventListener('offline', visibility);
		return () => {
			isAlive = false; stopFlight();
			document.removeEventListener('visibilitychange', visibility);
			window.removeEventListener('online', visibility);
			window.removeEventListener('offline', visibility);
		};
	}, [owner, id, key, generation]);
	const visible = state.key === key ? state : {data: null, stale: false, error: '', loading: !!id};
	return {...visible, refresh: () => setGeneration(current => current + 1)};
}
