import {fireEvent, render, screen, waitFor} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {cancelJob, fetchJob} from './jobApi';
import {JobProgressPanel} from './JobProgressPanel';
import {aiProgress} from './testing/fixtures';

vi.mock('./jobApi', () => ({fetchJob: vi.fn(), cancelJob: vi.fn()}));
afterEach(() => {vi.resetAllMocks();});

it('shows execution separately from uncertain success and cancels only on explicit request', async () => {
 const job = aiProgress();
 job.counts = {total: 2, succeeded: 1, running: 1};
 job.items = [{...job.items[0], state: 'succeeded', outcome: 'unknown'}, {...job.items[0], id: '55555555-5555-4555-8555-555555555555', state: 'running'}];
 vi.mocked(fetchJob).mockResolvedValue(job);
 vi.mocked(cancelJob).mockResolvedValue({canceled: true});
 render(<JobProgressPanel owner={'owner'} id={job.id} onRerunAction={vi.fn()} />);
 expect(await screen.findByText('Succeeded · Location unknown')).toBeVisible();
 expect(screen.getByText(/Reported tokens: Unknown/)).toBeVisible();
 expect(screen.getByText(/cannot recall/)).toBeVisible();
 expect(cancelJob).not.toHaveBeenCalled();
 fireEvent.click(screen.getByRole('button', {name: 'Cancel remaining work'}));
 await waitFor(() => expect(cancelJob).toHaveBeenCalledWith(job.id, expect.any(AbortSignal)));
 await waitFor(() => expect(fetchJob).toHaveBeenCalledTimes(2));
 expect(screen.getByText('Succeeded · Location unknown')).toBeVisible();
});

it('announces durable progress with a polite status region without moving keyboard focus', async () => {
 const job = aiProgress();
 vi.mocked(fetchJob).mockResolvedValue(job);
 render(<><button>{'Outside control'}</button><JobProgressPanel owner={'owner'} id={job.id} onRerunAction={vi.fn()} /></>);
 screen.getByRole('button', {name: 'Outside control'}).focus();
 await screen.findByText('Job in progress');
 expect(screen.getByRole('status')).toHaveTextContent('Queued: 1');
 expect(screen.getByRole('status')).toHaveAttribute('aria-live', 'polite');
 expect(screen.getByRole('button', {name: 'Outside control'})).toHaveFocus();
});

it('allows explicit cancellation of remaining queued work in a blocked job', async () => {
 const job = {...aiProgress(), blocked: true};
 vi.mocked(fetchJob).mockResolvedValue(job);
 vi.mocked(cancelJob).mockResolvedValue({canceled: true});
 render(<JobProgressPanel owner={'owner'} id={job.id} onRerunAction={vi.fn()} />);
 const cancel = await screen.findByRole('button', {name: 'Cancel remaining work'});
 expect(cancel).toBeEnabled();
 fireEvent.click(cancel);
 await waitFor(() => expect(cancelJob).toHaveBeenCalledTimes(1));
});

it('offers only failed members for retry and requires explicit member selection for reanalysis', async () => {
 const job = aiProgress();
 job.counts = {total: 2, failed: 1, succeeded: 1};
 job.items = [{...job.items[0], state: 'failed'}, {...job.items[0], id: '55555555-5555-4555-8555-555555555555', assetId: '66666666-6666-4666-8666-666666666666', state: 'succeeded', outcome: 'ambiguous'}];
 vi.mocked(fetchJob).mockResolvedValue(job);
 const rerun = vi.fn();
 render(<JobProgressPanel owner={'owner'} id={job.id} onRerunAction={rerun} />);
 fireEvent.click(await screen.findByRole('button', {name: 'Preview retry of failed assets'}));
 expect(rerun).toHaveBeenLastCalledWith({parentJobId: job.id, kind: 'retry-failed'}, [job.items[0].assetId]);
 expect(screen.getByRole('button', {name: 'Preview reanalysis of selected assets'})).toBeDisabled();
 fireEvent.click(screen.getByRole('checkbox', {name: `Reanalyze ${job.items[1].assetId}`}));
 fireEvent.click(screen.getByRole('button', {name: 'Preview reanalysis of selected assets'}));
 expect(rerun).toHaveBeenLastCalledWith({parentJobId: job.id, kind: 'reanalysis'}, [job.items[1].assetId]);
 expect(cancelJob).not.toHaveBeenCalled();
});
