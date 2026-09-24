import {act, fireEvent, render, screen} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {ResultDetail} from './ResultDetail';
import {fetchResult} from './resultsApi';
import {savedDraft} from './testing/draft';
import {resultDetail} from './testing/resultDetail';

vi.mock('./resultsApi', () => ({fetchResult: vi.fn()}));
afterEach(() => {vi.resetAllMocks(); vi.unstubAllGlobals();});
it('explicitly accepts a local draft and shows its saved revision without writing photos', async () => {
 const detail = resultDetail(); vi.mocked(fetchResult).mockResolvedValue(detail);
 const request = vi.fn().mockResolvedValue(new Response(JSON.stringify(savedDraft()), {status: 200})); vi.stubGlobal('fetch', request);
 render(<ResultDetail owner={'owner'} reference={{analysisId: detail.entry.analysisId!}} />);
 fireEvent.click(await screen.findByRole('button', {name: 'Accept as local draft'}));
 expect(await screen.findByText('Saved draft · Revision 1')).toBeVisible();
 expect(request.mock.calls.filter(([path]) => String(path).includes('/draft'))).toHaveLength(1);
 expect(request.mock.calls.some(([path]) => String(path).includes('/location'))).toBe(false);
});

it('saves zero camera coordinates with an exact revision through keyboard inputs', async () => {
 const detail = resultDetail(); vi.mocked(fetchResult).mockResolvedValue(detail);
 const request = vi.fn().mockImplementation(async (_path: string, init: RequestInit) => new Response(JSON.stringify(init.method === 'PATCH' ? {...savedDraft(), revision: 2, camera: {latitude: 0, longitude: 0}} : savedDraft()), {status: 200}));
 vi.stubGlobal('fetch', request);
 render(<ResultDetail owner={'owner'} reference={{analysisId: detail.entry.analysisId!}} />);
 fireEvent.click(await screen.findByRole('button', {name: 'Accept as local draft'}));
 fireEvent.change(await screen.findByLabelText('Camera latitude'), {target: {value: '0'}});
 fireEvent.change(screen.getByLabelText('Camera longitude'), {target: {value: '0'}});
 fireEvent.click(screen.getByRole('button', {name: 'Save draft'}));
 expect(await screen.findByText('Saved draft · Revision 2')).toBeVisible();
 const patch = request.mock.calls.find(([, init]) => init.method === 'PATCH')!;
 expect(new Headers(patch[1].headers).get('If-Match')).toBe('"1"');
 expect(JSON.parse(patch[1].body)).toMatchObject({camera: {latitude: 0, longitude: 0}});
});

it('preserves unsaved coordinates on conflict and reads saved state before allowing explicit replay', async () => {
 const detail = resultDetail(); vi.mocked(fetchResult).mockResolvedValue(detail);
 const request = vi.fn().mockImplementation(async (_path: string, init: RequestInit) => {
  if (init.method === 'PATCH') {return new Response(JSON.stringify({code: 'DRAFT_CONFLICT'}), {status: 412});}
  return new Response(JSON.stringify(init.method === 'GET' ? {...savedDraft(), revision: 2, camera: {latitude: 3, longitude: 4}} : savedDraft()), {status: 200});
 }); vi.stubGlobal('fetch', request);
 render(<ResultDetail owner={'owner'} reference={{analysisId: detail.entry.analysisId!}} />);
 fireEvent.click(await screen.findByRole('button', {name: 'Accept as local draft'}));
 fireEvent.change(await screen.findByLabelText('Camera latitude'), {target: {value: '7'}});
 fireEvent.change(screen.getByLabelText('Camera longitude'), {target: {value: '8'}});
 fireEvent.click(screen.getByRole('button', {name: 'Save draft'}));
 expect(await screen.findByText(/Another tab saved/)).toBeVisible();
 expect(screen.getByLabelText('Camera latitude')).toHaveValue(7);
 expect(screen.getByRole('button', {name: 'Save draft'})).toBeDisabled();
 fireEvent.click(screen.getByRole('button', {name: 'Compare saved revision'}));
 expect(await screen.findByText('Saved elsewhere: revision 2 · Camera 3, 4')).toBeVisible();
 expect(screen.getByLabelText('Camera latitude')).toHaveValue(7);
 expect(request.mock.calls.filter(([, init]) => init.method === 'PATCH')).toHaveLength(1);
 fireEvent.click(screen.getByRole('button', {name: 'Keep edits using saved revision'}));
 expect(screen.getByRole('button', {name: 'Save draft'})).toBeEnabled();
});

it('discards unsaved edits after an uncertain save that left the same revision', async () => {
 const detail = resultDetail(); vi.mocked(fetchResult).mockResolvedValue(detail);
 const request = vi.fn().mockImplementation(async (_path: string, init: RequestInit) => {
  if (init.method === 'PATCH') {throw new Error('lost connection');}
  return new Response(JSON.stringify(savedDraft()), {status: 200});
 }); vi.stubGlobal('fetch', request);
 render(<ResultDetail owner={'owner'} reference={{analysisId: detail.entry.analysisId!}} />);
 fireEvent.click(await screen.findByRole('button', {name: 'Accept as local draft'}));
 fireEvent.change(await screen.findByLabelText('Camera latitude'), {target: {value: '7'}});
 fireEvent.change(screen.getByLabelText('Camera longitude'), {target: {value: '8'}});
 fireEvent.click(screen.getByRole('button', {name: 'Save draft'}));
 expect(await screen.findByText(/Save outcome unknown/)).toBeVisible();
 fireEvent.click(screen.getByRole('button', {name: 'Compare saved revision'}));
 fireEvent.click(await screen.findByRole('button', {name: 'Discard edits and use saved revision'}));
 expect(screen.getByLabelText('Camera latitude')).toHaveValue(null);
 expect(screen.getByLabelText('Camera longitude')).toHaveValue(null);
 expect(request.mock.calls.filter(([, init]) => init.method === 'PATCH')).toHaveLength(1);
});


it('fences an unacknowledged acceptance across owner and result changes', async () => {
 const detail = resultDetail(); vi.mocked(fetchResult).mockResolvedValue(detail);
 let reply: (value: Response) => void = () => {};
 vi.stubGlobal('fetch', vi.fn(async () => new Promise<Response>(resolve => {reply = resolve;})));
 const view = render(<ResultDetail owner={'one'} reference={{analysisId: detail.entry.analysisId!}} />);
 fireEvent.click(await screen.findByRole('button', {name: 'Accept as local draft'}));
 const next = {...detail, entry: {...detail.entry, analysisId: detail.entry.jobId, label: 'Different private result'}};
 vi.mocked(fetchResult).mockResolvedValue(next);
 view.rerender(<ResultDetail owner={'two'} reference={{analysisId: next.entry.analysisId!}} />);
 expect(await screen.findByText('Different private result')).toBeVisible();
 await act(async () => reply(new Response(JSON.stringify(savedDraft()))));
 expect(screen.queryByText('Saved draft · Revision 1')).not.toBeInTheDocument();
 expect(screen.queryByLabelText('Camera latitude')).not.toBeInTheDocument();
 expect(screen.getByRole('button', {name: 'Accept as local draft'})).toBeEnabled();
});
