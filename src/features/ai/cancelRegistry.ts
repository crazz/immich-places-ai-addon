export function createCancelRegistry(): {
	register: (cancel: () => void) => () => void;
	cancelAll: () => void;
} {
	const cancels = new Set<() => void>();
	return {
		register(cancel: () => void) {
			cancels.add(cancel);
			return () => {
				cancels.delete(cancel);
			};
		},
		cancelAll() {
			for (const cancel of [...cancels]) {
				cancel();
			}
		}
	};
}
