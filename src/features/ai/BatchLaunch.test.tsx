import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import {BatchLaunch} from './BatchLaunch';
import {fetchProviders} from './providerApi';
import {previewSelection} from './selectionApi';
import {aiPreview, aiProfile} from './testing/fixtures';

vi.mock('./providerApi', () => ({fetchProviders: vi.fn()}));
vi.mock('./selectionApi', () => ({previewSelection: vi.fn()}));
afterEach(() => vi.resetAllMocks());

it('shows the authoritative scope filters before disclosure consent', async () => {
 const user = userEvent.setup();
 const preview = aiPreview();
 preview.scope = {view: 'album', gpsFilter: 'no-gps', hiddenFilter: 'visible', albumID: 'approved-album', folderPath: '/approved-folder', tagID: 'approved-tag', startDate: '2026-01-01', endDate: '2026-02-01'};
 vi.mocked(previewSelection).mockResolvedValue(preview);
 vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: []});
 render(<BatchLaunch input={{scope: {view: 'all', gpsFilter: 'all', hiddenFilter: 'visible'}, selected: preview.assetIDs, page: [], blockedReason: ''}} onSubmittedAction={() => undefined} />);
 await user.click(screen.getByRole('button', {name: 'Preview selected assets'}));
 expect(await screen.findByText('GPS: no-gps · Visibility: visible')).toBeVisible();
 expect(screen.getByText('Album: approved-album')).toBeVisible();
 expect(screen.getByText('Folder: /approved-folder')).toBeVisible();
 expect(screen.getByText('Tag: approved-tag')).toBeVisible();
 expect(screen.getByText('From: 2026-01-01')).toBeVisible();
 expect(screen.getByText('Through: 2026-02-01')).toBeVisible();
});

it('previews selected, page and all matching with authoritative exclusions and expiry', async () => {
 const user = userEvent.setup();
 const scope = {view: 'all' as const, gpsFilter: 'no-gps', hiddenFilter: 'visible'};
 const input = {scope, selected: ['selected'], page: ['page-a', 'page-b'], blockedReason: ''};
 vi.mocked(previewSelection).mockResolvedValue({mode: 'explicit', scope, snapshotID: null, policyVersion: 'selection-v1', assetIDs: [], requestedCount: 2, uniqueCount: 2, duplicateCount: 0, eligibleCount: 0, excludedCount: 2, exclusions: [{assetID: 'selected', reason: 'ineligible'}], createdAt: new Date().toISOString(), expiresAt: new Date(Date.now() + 60000).toISOString()});
 render(<BatchLaunch input={input} onSubmittedAction={() => undefined} />);
 await user.click(screen.getByRole('button', {name: 'Preview selected assets'}));
 expect(previewSelection).toHaveBeenLastCalledWith({mode: 'explicit', scope, assetIDs: ['selected']}, expect.any(AbortSignal));
 expect(await screen.findByText(/Eligible: 0/)).toBeVisible();
 expect(screen.getByText(/Excluded: 2/)).toBeVisible();
 await user.click(screen.getByRole('button', {name: 'Preview current page'}));
 expect(previewSelection).toHaveBeenLastCalledWith({mode: 'explicit', scope, assetIDs: ['page-a', 'page-b']}, expect.any(AbortSignal));
 await user.click(screen.getByRole('button', {name: 'Preview all matching assets'}));
 expect(previewSelection).toHaveBeenLastCalledWith({mode: 'all-matching', scope}, expect.any(AbortSignal));
 expect(screen.queryByRole('button', {name: 'Start analysis'})).not.toBeInTheDocument();
});


it('invalidates the displayed launch when selection or page inputs change', async () => {
 const user = userEvent.setup();
 const preview = aiPreview();
 const input = {scope: preview.scope, selected: preview.assetIDs, page: preview.assetIDs, blockedReason: ''};
 vi.mocked(previewSelection).mockResolvedValue(preview);
 vi.mocked(fetchProviders).mockResolvedValue({enabled: true, items: [aiProfile]});
 const {rerender} = render(<BatchLaunch input={input} onSubmittedAction={() => undefined} />);
 await user.click(screen.getByRole('button', {name: 'Preview selected assets'}));
 await screen.findByRole('button', {name: 'Start analysis'});
 expect(screen.getByRole('button', {name: 'Start analysis'})).toBeEnabled();
 rerender(<BatchLaunch input={{...input, page: ['different-page']}} onSubmittedAction={() => undefined} />);
 expect(screen.queryByLabelText(/I consent to sending/)).not.toBeInTheDocument();
 expect(screen.queryByRole('button', {name: 'Start analysis'})).not.toBeInTheDocument();
 expect(previewSelection).toHaveBeenCalledTimes(1);
});
