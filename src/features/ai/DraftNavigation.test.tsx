import {fireEvent, render, screen} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {AIResultsWorkspace} from './AIResultsWorkspace';
import {fetchResult, fetchResults} from './resultsApi';
import {savedDraft} from './testing/draft';
import {resultDetail} from './testing/resultDetail';

vi.mock('./resultsApi', () => ({fetchResult: vi.fn(), fetchResults: vi.fn()}));
afterEach(() => {vi.resetAllMocks(); vi.unstubAllGlobals(); window.history.replaceState(null, '', '/');});
it('offers keep editing or explicit discard before leaving unsaved review', async () => {
 const detail = resultDetail(); vi.mocked(fetchResult).mockResolvedValue(detail); vi.mocked(fetchResults).mockResolvedValue({items: [detail.entry]});
 const request = vi.fn().mockResolvedValue(new Response(JSON.stringify(savedDraft()), {status: 200})); vi.stubGlobal('fetch', request);
 render(<AIResultsWorkspace owner={'owner'} />);
 fireEvent.click(screen.getByRole('button', {name: 'AI Results'}));
 fireEvent.click(await screen.findByRole('button', {name: `Open result ${detail.entry.id}`}));
 fireEvent.click(await screen.findByRole('button', {name: 'Accept as local draft'}));
 fireEvent.change(await screen.findByLabelText('Camera latitude'), {target: {value: '7'}});
 expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
 fireEvent.click(screen.getByRole('button', {name: 'Back to results'}));
 expect(await screen.findByRole('alertdialog', {name: 'Unsaved draft edits'})).toBeVisible();
 fireEvent.click(screen.getByRole('button', {name: 'Keep editing'}));
 expect(screen.getByLabelText('Camera latitude')).toHaveValue(7);
 fireEvent.click(screen.getByRole('button', {name: 'Close dialog'}));
 fireEvent.click(screen.getByRole('button', {name: 'Discard unsaved edits'}));
 expect(screen.queryByLabelText('Camera latitude')).not.toBeInTheDocument();
 expect(request.mock.calls.filter(([, init]) => init.method === 'PATCH')).toHaveLength(0);
});
