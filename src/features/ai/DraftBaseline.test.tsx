import {fireEvent, render, screen, waitFor} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {ResultDetail} from './ResultDetail';
import {fetchResult} from './resultsApi';
import {savedDraft} from './testing/draft';
import {resultDetail} from './testing/resultDetail';

vi.mock('./resultsApi', () => ({fetchResult: vi.fn()}));
afterEach(() => {vi.resetAllMocks(); vi.unstubAllGlobals();});
it('shows exact partial baseline and requires explicit current-image acknowledgement', async () => {
 const detail = resultDetail(); vi.mocked(fetchResult).mockResolvedValue(detail);
 vi.stubGlobal('URL', class extends URL {static createObjectURL = (): string => 'blob:baseline'; static revokeObjectURL = vi.fn();});
 const baseline = {...savedDraft().baseline, status: 'reviewed', latitude: 0, imageIdentity: `v1:${'a'.repeat(64)}`, sourceDigest: 'c'.repeat(64), assetId: detail.entry.assetId, ownerId: detail.entry.jobId, checksum: 'checksum', type: 'IMAGE', observedAt: '2026-09-24T12:00:00Z'};
 const observation = {id: detail.entry.id, draftId: savedDraft().id, revision: 1, baseline, expiresAt: '2026-09-24T12:05:00Z', previewUrl: `/ai/jobs/${detail.entry.jobId}/items/${detail.entry.id}/thumbnail`, originalSourceMatches: false};
 vi.stubGlobal('fetch', vi.fn(async (path: string, init: RequestInit) => {
  if (path.endsWith('/thumbnail')) {return new Response(new Uint8Array([1, 2, 3]), {headers: new Headers([['Content-Type', 'image/jpeg']])});}
  const body = init.body ? JSON.parse(String(init.body)) : {};
  const response = path.endsWith('/baseline') ? body.observationId ? {...savedDraft(), revision: 2, baseline} : observation : savedDraft();
  return new Response(JSON.stringify(response), {status: 200});
 }));
 render(<ResultDetail owner={'owner'} reference={{analysisId: detail.entry.analysisId!}} />);
 fireEvent.click(await screen.findByRole('button', {name: 'Accept as local draft'}));
 fireEvent.click(await screen.findByRole('button', {name: 'Review current source and GPS'}));
 expect(await screen.findByText('Observed GPS: latitude 0 · longitude absent')).toBeVisible();
 expect(screen.getByText(/Original source fingerprint differs/)).toBeVisible();
 expect(screen.getByRole('button', {name: 'Acknowledge displayed baseline'})).toBeDisabled();
 fireEvent.load(await screen.findByRole('img', {name: 'Current source photo'}));
 await waitFor(() => expect(screen.getByLabelText('I reviewed this current image and its displayed GPS')).toBeEnabled());
 fireEvent.click(screen.getByLabelText('I reviewed this current image and its displayed GPS'));
 fireEvent.click(screen.getByRole('button', {name: 'Acknowledge displayed baseline'}));
 expect(await screen.findByText('Saved draft · Revision 2')).toBeVisible();
});
