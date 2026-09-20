import {act, renderHook} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {fetchJob} from './jobApi';
import {aiProgress} from './testing/fixtures';
import {useJobProgress} from './useJobProgress';

vi.mock('./jobApi', () => ({fetchJob: vi.fn()}));
afterEach(() => {vi.restoreAllMocks(); vi.resetAllMocks(); vi.useRealTimers();});

it('polls active durable jobs every two seconds and stops after terminal completion without submitting', async () => {
 vi.useFakeTimers();
 const queued = aiProgress();
 const completed = {...queued, counts: {succeeded: 1, total: 1}, items: queued.items.map(item => ({...item, state: 'succeeded' as const, outcome: 'unknown' as const}))};
 vi.mocked(fetchJob).mockResolvedValueOnce(queued).mockResolvedValue(completed);
 const {result} = renderHook(() => useJobProgress('owner', queued.id));
 await act(async () => {await Promise.resolve();});
 expect(result.current.data?.id).toBe(queued.id);
 await act(async () => {await vi.advanceTimersByTimeAsync(2000);});
 expect(result.current.data?.items[0].outcome).toBe('unknown');
 await act(async () => {await vi.advanceTimersByTimeAsync(60000);});
 expect(fetchJob).toHaveBeenCalledTimes(2);
});


it('pauses hidden and offline polling and immediately refreshes when visible and online', async () => {
 vi.useFakeTimers();
 let isHidden = false;
 let isOnline = true;
 vi.spyOn(document, 'hidden', 'get').mockImplementation(() => isHidden);
 vi.spyOn(navigator, 'onLine', 'get').mockImplementation(() => isOnline);
 const job = aiProgress();
 vi.mocked(fetchJob).mockResolvedValue(job);
 const {result} = renderHook(() => useJobProgress('owner', job.id));
 await act(async () => {await Promise.resolve();});
 await act(async () => {isHidden = true; document.dispatchEvent(new Event('visibilitychange')); await vi.advanceTimersByTimeAsync(60000);});
 expect(fetchJob).toHaveBeenCalledTimes(1);
 expect(result.current.stale).toBe(true);
 await act(async () => {isHidden = false; isOnline = false; window.dispatchEvent(new Event('offline')); await vi.advanceTimersByTimeAsync(60000);});
 expect(result.current.error).toContain('Offline');
 expect(fetchJob).toHaveBeenCalledTimes(1);
 await act(async () => {isOnline = true; window.dispatchEvent(new Event('online')); await Promise.resolve();});
 expect(fetchJob).toHaveBeenCalledTimes(2);
 expect(result.current.stale).toBe(false);
});

it('clears private state and ignores an old response after switching owner or job', async () => {
 const first = Promise.withResolvers<ReturnType<typeof aiProgress>>();
 const old = aiProgress();
 const current = {...old, id: '44444444-4444-4444-8444-444444444444', counts: {failed: 1, total: 1}};
 vi.mocked(fetchJob).mockReturnValueOnce(first.promise).mockResolvedValue(current);
 const {result, rerender} = renderHook(({owner, id}) => useJobProgress(owner, id), {initialProps: {owner: 'old-owner', id: old.id}});
 const oldSignal = vi.mocked(fetchJob).mock.calls[0][1];
 rerender({owner: 'new-owner', id: current.id});
 expect(result.current.data).toBeNull();
 await act(async () => {await Promise.resolve();});
 expect(result.current.data?.id).toBe(current.id);
 await act(async () => {first.resolve(old); await first.promise;});
 expect(oldSignal?.aborted).toBe(true);
 expect(result.current.data?.id).toBe(current.id);
});

it('backs off failed reads to thirty seconds and lets an explicit refresh recover', async () => {
 vi.useFakeTimers();
 vi.mocked(fetchJob).mockRejectedValue(new Error('Synthetic network failure'));
 const job = aiProgress();
 const {result} = renderHook(() => useJobProgress('owner', job.id));
 await act(async () => {await Promise.resolve();});
 expect(result.current.stale).toBe(true);
 for (const delay of [4000, 8000, 16000, 30000, 30000]) {
  const before = vi.mocked(fetchJob).mock.calls.length;
  await act(async () => {await vi.advanceTimersByTimeAsync(delay - 1);});
  expect(fetchJob).toHaveBeenCalledTimes(before);
  await act(async () => {await vi.advanceTimersByTimeAsync(1);});
  expect(fetchJob).toHaveBeenCalledTimes(before + 1);
 }
 vi.mocked(fetchJob).mockResolvedValue({...job, blocked: true});
 await act(async () => {result.current.refresh();});
 expect(result.current.stale).toBe(false);
 expect(result.current.data?.blocked).toBe(true);
 const finalCount = vi.mocked(fetchJob).mock.calls.length;
 await act(async () => {await vi.advanceTimersByTimeAsync(60000);});
 expect(fetchJob).toHaveBeenCalledTimes(finalCount);
});
