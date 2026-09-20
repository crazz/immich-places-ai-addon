import {fireEvent, render, screen} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {previewSelection} from '@/features/ai/selectionApi';
import {PhotoListContainer} from '@/shared/components/PhotoListContainer';

const state = vi.hoisted(() => ({gpxStatusFilter: 'all', viewMode: 'album', selectedAlbumID: '33333333-3333-4333-8333-333333333333', currentPage: 2, isLoadingAssets: false}));
vi.mock('@/features/auth/AuthContext', () => ({useAuth: () => ({user: {ID: 'owner'}, mapMarkerCount: 0})}));
vi.mock('@/features/auth/UserMenu', () => ({UserMenu: () => null}));
vi.mock('@/features/filterBar/useMissingLocationCount', () => ({useMissingLocationCount: () => 0}));
vi.mock('@/features/gpxImport/GPXImportContext', () => ({useGPXImportContext: () => ({step: 'idle', previews: []})}));
vi.mock('@/features/selection/useDiscardPendingLocations', () => ({useDiscardPendingLocations: () => vi.fn()}));
vi.mock('@/features/selection/selectionStateHelpers', () => ({deriveAlreadyAppliedIDs: () => new Set()}));
vi.mock('@/shared/context/AppContext', () => ({
 useBackend: () => ({}),
useView: () => ({...state, gpsFilter: 'no-gps', hiddenFilter: 'visible', selectedTagID: 'tag', startDate: '2026-08-01', endDate: '2026-08-03'}),
 useCatalog: () => ({albums: [], assets: [{immichID: '11111111-1111-4111-8111-111111111111'}], currentPage: state.currentPage, isLoadingAssets: state.isLoadingAssets}),
 useSelection: () => ({selectedAssets: [], gpxStatusFilter: state.gpxStatusFilter}),
useUIMap: () => ({})
}));
vi.mock('@/shared/components/PhotoList', () => ({PhotoList: ({view}: {view: {trailingAction: React.ReactNode}}) => <div>{view.trailingAction}</div>}));
vi.mock('@/features/ai/selectionApi', () => ({previewSelection: vi.fn()}));
afterEach(() => {vi.resetAllMocks(); Object.assign(state, {gpxStatusFilter: 'all', viewMode: 'album', selectedAlbumID: '33333333-3333-4333-8333-333333333333', isLoadingAssets: false});});

it('passes the exact visible gallery scope and current page to AI preview without dropping filters', async () => {
 vi.mocked(previewSelection).mockRejectedValue(new Error('Synthetic preview boundary reached'));
 render(<PhotoListContainer />);
 fireEvent.click(screen.getByRole('button', {name: 'AI Locate'}));
 fireEvent.click(screen.getByRole('button', {name: 'Preview current page'}));
 expect(await screen.findByText('Synthetic preview boundary reached')).toBeVisible();
 expect(previewSelection).toHaveBeenCalledWith({mode: 'explicit', scope: {view: 'album', albumID: state.selectedAlbumID, gpsFilter: 'no-gps', hiddenFilter: 'visible', tagID: 'tag', startDate: '2026-08-01', endDate: '2026-08-03'}, assetIDs: ['11111111-1111-4111-8111-111111111111']}, expect.any(AbortSignal));
});

it('blocks an unsupported GPX scope instead of silently broadening the preview', async () => {
 state.gpxStatusFilter = 'new';
 render(<PhotoListContainer />);
 fireEvent.click(screen.getByRole('button', {name: 'AI Locate'}));
 expect(screen.getByRole('button', {name: 'Preview all matching assets'})).toBeDisabled();
 expect(screen.getByText(/GPX preview or status filter/)).toBeVisible();
 expect(previewSelection).not.toHaveBeenCalled();
});

it('requires a settled concrete catalog scope before a new preview', () => {
 state.selectedAlbumID = '';
 const {rerender} = render(<PhotoListContainer />);
 fireEvent.click(screen.getByRole('button', {name: 'AI Locate'}));
 expect(screen.getByRole('button', {name: 'Preview current page'})).toBeDisabled();
 expect(screen.getByText(/Open a concrete album or folder/)).toBeVisible();
 state.viewMode = 'timeline'; state.isLoadingAssets = true;
 rerender(<PhotoListContainer />);
 expect(screen.getByRole('button', {name: 'Preview all matching assets'})).toBeDisabled();
 expect(screen.getByText(/Wait for the current catalog page/)).toBeVisible();
 expect(previewSelection).not.toHaveBeenCalled();
});
