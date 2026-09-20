'use client';

import {useCallback, useMemo} from 'react';

import {AIWorkspace} from '@/features/ai';
import {useAuth} from '@/features/auth/AuthContext';
import {UserMenu} from '@/features/auth/UserMenu';
import {useMissingLocationCount} from '@/features/filterBar/useMissingLocationCount';
import {useGPXImportContext} from '@/features/gpxImport/GPXImportContext';
import {deriveAlreadyAppliedIDs} from '@/features/selection/selectionStateHelpers';
import {useDiscardPendingLocations} from '@/features/selection/useDiscardPendingLocations';
import {PhotoList} from '@/shared/components/PhotoList';
import {useBackend, useCatalog, useSelection, useUIMap, useView} from '@/shared/context/AppContext';

import type {TAlbumRow} from '@/shared/types/album';
import type {TViewMode} from '@/shared/types/view';
import type {ReactElement} from 'react';

export function PhotoListContainer(): ReactElement {
	const {isSyncing, syncError, resyncAction} = useBackend();
	const {
		gpsFilter,
		setGPSFilterAction,
		hiddenFilter,
		setHiddenFilterAction,
		pageSize,
		setPageSizeAction,
		gridColumns,
		setGridColumnsAction,
		visibleMarkerLimit,
		setVisibleMarkerLimitAction,
		viewMode,
		setViewModeAction,
		selectedAlbumID,
		selectAlbumAction,
		selectedFolderPath,
		selectFolderAction,
		selectedTagID,
		selectTagAction,
		startDate,
		endDate,
		setDateRangeAction
	} = useView();
	const {
		albums,
		isLoadingAlbums,
		albumsError,
		assets,
		total,
		currentPage,
		isLoadingAssets,
		assetsError,
		loadPageAction
	} = useCatalog();
	const {user, mapMarkerCount} = useAuth();
	const {clearSelectionAction, selectedAssets, pendingLocationsByAssetID, gpxStatusFilter, setGPXStatusFilterAction} =
		useSelection();
	const {closeLightboxAction} = useUIMap();

	const {
		step: gpxStep,
		isLoading: isGPXLoading,
		error: gpxError,
		previews: gpxPreviews,
		uploadAndPreview: gpxUploadAndPreview,
		setPreviews: gpxSetPreviews,
		reset: gpxReset
	} = useGPXImportContext();

	const isGPXPanelActive = gpxStep === 'preview' && gpxPreviews.length > 0;

	const selectedAlbum = useMemo<TAlbumRow | null>(
		() => albums.find(album => album.immichID === selectedAlbumID) ?? null,
		[albums, selectedAlbumID]
	);

	const missingCount = useMissingLocationCount();
	const selectedIDs = useMemo(() => new Set(selectedAssets.map(a => a.immichID)), [selectedAssets]);
	const alreadyAppliedIDs = useMemo(
		() => deriveAlreadyAppliedIDs(pendingLocationsByAssetID),
		[pendingLocationsByAssetID]
	);
	const closeAlbumAndClearPending = useDiscardPendingLocations();

	const handleToggleViewMode = (mode: TViewMode): void => {
		if (mode === viewMode) {
			return;
		}
		if (selectedAlbumID || selectedFolderPath) {
			closeAlbumAndClearPending(() => {
				closeLightboxAction();
				selectAlbumAction(null);
				selectFolderAction(null);
				setViewModeAction(mode);
			});
			return;
		}
		closeLightboxAction();
		if (mode === 'album') {
			selectTagAction(null);
		}
		setViewModeAction(mode);
	};

	const handleBackToAlbums = (): void => {
		if (!selectedAlbumID) {
			return;
		}
		closeAlbumAndClearPending(() => {
			closeLightboxAction();
			selectAlbumAction(null);
		});
	};

	const handleBackToFolders = (): void => {
		if (!selectedFolderPath) {
			return;
		}
		closeAlbumAndClearPending(() => {
			closeLightboxAction();
			selectFolderAction(null);
		});
	};

	const handleSelectAlbum = (albumID: string): void => {
		if (selectedAlbumID === albumID) {
			return;
		}
		if (selectedAlbumID) {
			closeAlbumAndClearPending(() => {
				closeLightboxAction();
				selectAlbumAction(albumID);
			});
			return;
		}
		closeLightboxAction();
		selectAlbumAction(albumID);
	};

	const handleSelectFolder = (folderPath: string): void => {
		if (selectedFolderPath === folderPath) {
			return;
		}
		if (selectedFolderPath) {
			closeAlbumAndClearPending(() => {
				closeLightboxAction();
				selectFolderAction(folderPath);
			});
			return;
		}
		closeLightboxAction();
		selectFolderAction(folderPath);
	};

	const handleLoadPageAction = async (page: number): Promise<void> => {
		clearSelectionAction();
		return loadPageAction(page);
	};

	const handleGPXCancel = useCallback((): void => {
		closeAlbumAndClearPending(gpxReset);
	}, [closeAlbumAndClearPending, gpxReset]);

	const handleGPXAutoReset = useCallback((): void => {
		clearSelectionAction();
		gpxReset();
	}, [clearSelectionAction, gpxReset]);

	let activeGPXPreviews: typeof gpxPreviews = [];
	if (isGPXPanelActive) {
		activeGPXPreviews = gpxPreviews;
	}
	let aiBlockedReason = '';
	if (isGPXPanelActive || gpxStatusFilter !== 'all') {
		aiBlockedReason = 'Close the GPX preview or status filter before creating an AI selection.';
	} else if ((viewMode === 'album' && !selectedAlbumID) || (viewMode === 'folders' && !selectedFolderPath)) {
		aiBlockedReason = 'Open a concrete album or folder, or use Timeline before creating an AI selection.';
	} else if (isLoadingAssets || assetsError || isSyncing) {
		aiBlockedReason = 'Wait for the current catalog page to load successfully before creating an AI selection.';
	}

	let gpxImportProp:
		| {
				uploadAndPreview: (files: File[], maxGapSeconds?: number) => Promise<void>;
				setPreviews: (previews: typeof gpxPreviews) => void;
				isLoading: boolean;
				error: string | null;
		  }
		| undefined;
	if (!isGPXPanelActive) {
		gpxImportProp = {
			uploadAndPreview: gpxUploadAndPreview,
			setPreviews: gpxSetPreviews,
			isLoading: isGPXLoading,
			error: gpxError
		};
	}

	return (
		<PhotoList
			backend={{
				isSyncing,
				syncError,
				onResyncAction: resyncAction
			}}
			view={{
				gpsFilter,
				hiddenFilter,
				pageSize,
				gridColumns,
				visibleMarkerLimit,
				visibleMarkerTotalCount: mapMarkerCount,
				viewMode,
				selectedAlbumID,
				selectedAlbum,
				selectedFolderPath,
				missingCount,
				onGPSFilterAction: setGPSFilterAction,
				onHiddenFilterAction: setHiddenFilterAction,
				onPageSizeAction: setPageSizeAction,
				onGridColumnsAction: setGridColumnsAction,
				onVisibleMarkerLimitAction: setVisibleMarkerLimitAction,
				onViewModeAction: handleToggleViewMode,
				onBackToAlbumsAction: handleBackToAlbums,
				onBackToFoldersAction: handleBackToFolders,
				gpxPreviews: activeGPXPreviews,
				gpxError,
				startDate,
				endDate,
				onDateRangeAction: setDateRangeAction,
				selectedTagID,
				onTagAction: selectTagAction,
				onGPXResetAction: handleGPXAutoReset,
				onGPXCancelAction: handleGPXCancel,
				trailingAction: <>
					{user && <AIWorkspace
owner={user.ID} input={{
						scope: {
							view: selectedAlbumID ? 'album' : selectedFolderPath ? 'folder' : 'all',
							gpsFilter,
hiddenFilter,
							...(selectedAlbumID ? {albumID: selectedAlbumID} : {}),
							...(selectedFolderPath ? {folderPath: selectedFolderPath} : {}),
							...(selectedTagID ? {tagID: selectedTagID} : {}),
							...(startDate ? {startDate} : {}),
...(endDate ? {endDate} : {})
						},
						selected: selectedAssets.map(asset => asset.immichID),
page: assets.map(asset => asset.immichID),
						revisionKey: `${currentPage}:${pageSize}`,
						blockedReason: aiBlockedReason
					}} />}
					<UserMenu gpxImport={gpxImportProp} />
                    </>,
				gpxStatusFilter,
				onGPXStatusFilterAction: setGPXStatusFilterAction
			}}
			catalog={{
				assets,
				total,
				currentPage,
				isLoadingAlbums,
				albumsError,
				isLoadingAssets,
				assetsError,
				onLoadPageAction: handleLoadPageAction,
				onSelectAlbumAction: handleSelectAlbum,
				onSelectFolderAction: handleSelectFolder,
				onRetrySyncAction: resyncAction
			}}
			selection={{
				selectedIDs,
				alreadyAppliedIDs
			}}
		/>
	);
}
