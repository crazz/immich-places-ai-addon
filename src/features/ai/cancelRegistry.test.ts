import {expect, it, vi} from 'vitest';

import {createCancelRegistry} from './cancelRegistry';

it('invokes every registered cancel on cancelAll and supports unregister', () => {
	const registry = createCancelRegistry();
	const first = vi.fn();
	const second = vi.fn();
	const unregisterFirst = registry.register(first);
	registry.register(second);
	registry.cancelAll();
	expect(first).toHaveBeenCalledTimes(1);
	expect(second).toHaveBeenCalledTimes(1);
	unregisterFirst();
	registry.cancelAll();
	expect(first).toHaveBeenCalledTimes(1);
	expect(second).toHaveBeenCalledTimes(2);
});
