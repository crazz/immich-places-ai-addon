import {fireEvent, render, screen, waitFor} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {AIWorkspace} from './AIWorkspace';
import {cancelJob, fetchJob, submitJob} from './jobApi';
import {aiProgress} from './testing/fixtures';

vi.mock('./jobApi', () => ({fetchJob: vi.fn(), fetchJobs: vi.fn(), cancelJob: vi.fn(), submitJob: vi.fn()}));
afterEach(() => {vi.resetAllMocks(); window.history.replaceState(null, '', '/');});

it('restores a job from navigation and closes observation without canceling or submitting', async () => {
 const job = aiProgress();
 window.history.replaceState(null, '', `/?aiJob=${job.id}`);
 vi.mocked(fetchJob).mockResolvedValue(job);
 render(<AIWorkspace owner={'owner'} input={{scope: {view: 'all', gpsFilter: 'all', hiddenFilter: 'visible'}, selected: [], page: [], blockedReason: ''}} />);
 expect(await screen.findByText(`Job ${job.id}`)).toBeVisible();
 await waitFor(() => expect(fetchJob).toHaveBeenCalledTimes(1));
 const signal = vi.mocked(fetchJob).mock.calls[0][1];
 fireEvent.click(screen.getByRole('button', {name: 'Close dialog'}));
 expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
 expect(signal?.aborted).toBe(true);
 expect(window.location.search).toBe('');
 expect(cancelJob).not.toHaveBeenCalled();
 expect(submitJob).not.toHaveBeenCalled();
});
