import {act, renderHook, waitFor} from '@testing-library/react';
import {expect, it, vi} from 'vitest';

import {useResultRead} from './useResultRead';

it('retains stale reads on transient failure but discards late responses and private data on owner changes', async () => {
 let resolve: (value: string) => void = () => {};
 const load = vi.fn<(signal: AbortSignal) => Promise<string>>().mockResolvedValueOnce('saved').mockRejectedValueOnce(new Error('offline')).mockImplementationOnce(async () => new Promise(done => {resolve = done;})).mockResolvedValue('new owner');
 const view = renderHook(({owner}) => useResultRead(owner, 'page', load), {initialProps: {owner: 'one'}});
 await waitFor(() => expect(view.result.current.data).toBe('saved'));
 act(() => view.result.current.refresh());
 await waitFor(() => expect(view.result.current.error).not.toBe(''));
 expect(view.result.current.data).toBe('saved');
 act(() => view.result.current.refresh());
 await waitFor(() => expect(load).toHaveBeenCalledTimes(3));
 view.rerender({owner: 'two'});
 expect(view.result.current.data).toBeNull();
 await act(async () => resolve('old owner'));
 await waitFor(() => expect(view.result.current.data).toBe('new owner'));
 expect(load.mock.calls[2][0].aborted).toBe(true);
 expect(load).toHaveBeenCalledTimes(4);
});
