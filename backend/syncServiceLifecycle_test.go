package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestDoFullSync(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	factory, immich := newFullMockImmichFactory(t)
	nom := newMockNominatimServer(t)
	svc := newSyncService(db, factory, nom)
	t.Cleanup(svc.wg.Wait)

	svc.doUserFullSync(ctx, testUserID, immich)

	total, _ := db.countAssets(ctx, testUserID)
	if total != 1 {
		t.Errorf("expected 1 asset after full sync, got %d", total)
	}

	lastSync, _ := db.getSyncState(ctx, testUserID, "lastSyncAt")
	if lastSync == nil {
		t.Error("expected lastSyncAt to be set")
	}
	lastFullSync, _ := db.getSyncState(ctx, testUserID, "lastFullSyncAt")
	if lastFullSync == nil {
		t.Error("expected lastFullSyncAt to be set")
	}
	backfillDone, _ := db.getSyncState(ctx, testUserID, "libraryIDBackfillDone")
	if backfillDone == nil || *backfillDone != "true" {
		t.Errorf("expected libraryIDBackfillDone=true, got %v", backfillDone)
	}
}

func TestDoFullSyncDoesNotMarkBackfillDoneWhenLibrarySyncFails(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	factory, immich := newMockImmichFactoryNoRetry(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/search/metadata":
			json.NewEncoder(w).Encode(ImmichSearchResponse{
				Assets: struct {
					Items    []ImmichAssetResponse `json:"items"`
					NextPage *string               `json:"nextPage"`
				}{Items: []ImmichAssetResponse{}},
			})
		case r.URL.Path == "/api/stacks":
			json.NewEncoder(w).Encode([]ImmichStackResponse{})
		case r.URL.Path == "/api/libraries":
			w.WriteHeader(http.StatusForbidden)
		case r.URL.Path == "/api/albums":
			json.NewEncoder(w).Encode([]ImmichAlbumResponse{})
		default:
			http.NotFound(w, r)
		}
	})
	nom := newMockNominatimServer(t)
	svc := newSyncService(db, factory, nom)
	t.Cleanup(svc.wg.Wait)

	svc.doUserFullSync(ctx, testUserID, immich)

	backfillDone, _ := db.getSyncState(ctx, testUserID, "libraryIDBackfillDone")
	if backfillDone != nil {
		t.Errorf("expected libraryIDBackfillDone to remain unset when library sync fails, got %v", backfillDone)
	}
}

func TestDoIncrementalSyncFallsBackToFull(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	factory, immich := newFullMockImmichFactory(t)
	nom := newMockNominatimServer(t)
	svc := newSyncService(db, factory, nom)
	t.Cleanup(svc.wg.Wait)

	svc.doUserIncrementalSync(ctx, testUserID, immich)

	lastFullSync, _ := db.getSyncState(ctx, testUserID, "lastFullSyncAt")
	if lastFullSync == nil {
		t.Error("expected full sync fallback when no lastSyncAt exists")
	}
}

func TestDoIncrementalSync(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	factory, immich := newFullMockImmichFactory(t)
	nom := newMockNominatimServer(t)
	svc := newSyncService(db, factory, nom)
	t.Cleanup(svc.wg.Wait)

	db.setSyncState(ctx, testUserID, "lastSyncAt", "2024-01-01T00:00:00Z")

	svc.doUserIncrementalSync(ctx, testUserID, immich)

	total, _ := db.countAssets(ctx, testUserID)
	if total != 1 {
		t.Errorf("expected 1 asset after incremental sync, got %d", total)
	}
}

func TestDoIncrementalSyncForcesFullWhenBackfillNeeded(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	factory, immich := newFullMockImmichFactory(t)
	nom := newMockNominatimServer(t)
	svc := newSyncService(db, factory, nom)
	t.Cleanup(svc.wg.Wait)

	db.setSyncState(ctx, testUserID, "lastSyncAt", "2024-01-01T00:00:00Z")
	db.setSyncState(ctx, testUserID, "hasLibraryAccess", "true")
	db.upsertLibrary(ctx, "lib1", "External", 10)
	seedAsset(t, db, "a-existing", ptr(48.85), ptr(2.35), "2024-01-01T12:00:00Z")

	svc.doUserIncrementalSync(ctx, testUserID, immich)

	lastFullSync, _ := db.getSyncState(ctx, testUserID, "lastFullSyncAt")
	if lastFullSync == nil {
		t.Error("expected full sync when libraryID backfill is needed")
	}
}
