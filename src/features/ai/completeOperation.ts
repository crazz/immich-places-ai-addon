export async function withCompleteOperationDeadline<T>(
	timeoutMs: number,
	upstream: AbortSignal | undefined,
	run: (signal: AbortSignal) => Promise<T>
): Promise<T> {
	const controller = new AbortController();
	const abortFromUpstream = (): void => controller.abort();
	if (upstream) {
		if (upstream.aborted) {
			controller.abort();
		} else {
			upstream.addEventListener('abort', abortFromUpstream, {once: true});
		}
	}
	const timeoutID = setTimeout(() => controller.abort(), timeoutMs);
	try {
		return await run(controller.signal);
	} finally {
		clearTimeout(timeoutID);
		upstream?.removeEventListener('abort', abortFromUpstream);
	}
}

export async function raceAbort<T>(signal: AbortSignal, work: Promise<T>): Promise<T> {
	if (signal.aborted) {
		void work.catch(() => undefined);
		throw abortError();
	}
	return new Promise<T>((resolve, reject) => {
		const onAbort = (): void => {
			signal.removeEventListener('abort', onAbort);
			void work.catch(() => undefined);
			reject(abortError());
		};
		signal.addEventListener('abort', onAbort, {once: true});
		work.then(value => {
			signal.removeEventListener('abort', onAbort);
			resolve(value);
		}, failure => {
			signal.removeEventListener('abort', onAbort);
			reject(failure);
		});
	});
}

function abortError(): DOMException {
	return new DOMException('The operation was aborted.', 'AbortError');
}
