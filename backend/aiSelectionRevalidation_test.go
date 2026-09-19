package main

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"immich-places-backend/internal/ai/selection"
)

func selectionFixture(t *testing.T) (*Database, *aiSelectionStore, selection.Manifest) {
	t.Helper()
	db := newTestDB(t)
	seedAsset(t, db, selectionA, nil, nil, "2026-09-19")
	seedAsset(t, db, selectionB, nil, nil, "2026-09-19")
	store := selectionStore(t, db)
	result, err := store.preview(context.Background(), selection.Input{Mode: "explicit", AssetIDs: []string{selectionA, selectionB}, Scope: &selection.Scope{View: "all"}}, testUserID)
	if err != nil {
		t.Fatal(err)
	}
	return db, store, result
}

func TestAISelectionRetainedTargetChangesStaleWholeSnapshot(t *testing.T) {
	for _, query := range []string{
		"DELETE FROM assets WHERE immichID=?",
		"UPDATE assets SET isHidden=1 WHERE immichID=?",
		"UPDATE assets SET type='VIDEO' WHERE immichID=?",
		"UPDATE assets SET stackPrimaryAssetID='primary' WHERE immichID=?",
		"UPDATE assets SET latitude=1,longitude=2 WHERE immichID=?",
	} {
		t.Run(query, func(t *testing.T) {
			db, store, result := selectionFixture(t)
			selectionSQL(t, db, query, selectionB)
			loaded, err := store.load(context.Background(), testUserID, *result.SnapshotID)
			if err == nil || loaded.SnapshotID != nil {
				t.Fatalf("stale selection usable: %+v %v", loaded, err)
			}
		})
	}
}

func TestAISelectionAbsoluteExpiryAndPolicy(t *testing.T) {
	for _, kind := range []string{"expiry", "policy"} {
		t.Run(kind, func(t *testing.T) {
			db, store, result := selectionFixture(t)
			if kind == "expiry" {
				store.now = func() time.Time { return result.ExpiresAt }
			} else {
				result.PolicyVersion = "selection-future"
				data, err := json.Marshal(result)
				if err != nil {
					t.Fatal(err)
				}
				selectionSQL(t, db, "UPDATE ai_selection_snapshots SET manifest=?", string(data))
			}
			loaded, err := store.load(context.Background(), testUserID, *result.SnapshotID)
			if err == nil || loaded.SnapshotID != nil {
				t.Fatalf("expired/policy selection accepted: %+v %v", loaded, err)
			}
		})
	}
}

func TestAISelectionInstallationRotation(t *testing.T) {
	db, store, result := selectionFixture(t)
	if err := store.bind(context.Background(), "https://immich.example/api/", "1"); err != nil {
		t.Fatal(err)
	}
	if err := store.bind(context.Background(), "https://immich.example/api", "1"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.load(context.Background(), testUserID, *result.SnapshotID); err != nil {
		t.Fatalf("ordinary restart invalidated selection: %v", err)
	}
	if err := store.bind(context.Background(), "https://immich.example/api", "2"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.load(context.Background(), testUserID, *result.SnapshotID); err == nil {
		t.Fatal("old installation selection remains usable")
	}
	var n int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&n); err != nil || n != 0 {
		t.Fatalf("rotation retained snapshots: %d %v", n, err)
	}
}

func selectionStore(t *testing.T, db *Database) *aiSelectionStore {
	t.Helper()
	store := newAISelectionStore(db)
	store.enabled = true
	if err := store.bind(context.Background(), "https://immich.example/api", "1"); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestAISelectionRejectsInconsistentStoredManifest(t *testing.T) {
	db, store, result := selectionFixture(t)
	result.RequestedCount = 999
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, db, "UPDATE ai_selection_snapshots SET manifest=?", string(data))
	if loaded, err := store.load(context.Background(), testUserID, *result.SnapshotID); err == nil || loaded.SnapshotID != nil {
		t.Fatal("corrupt counts accepted")
	}
}

func TestAISelectionReopenDoesNotExpandOrRenew(t *testing.T) {
	db, store, result := selectionFixture(t)
	selectionSQL(t, db, "UPDATE assets SET syncedAt='new-sync',stackAssetCount=3,originalFileName='changed.jpg',city='unrelated'")
	seedAsset(t, db, selectionID(3), nil, nil, "2026-09-19")
	var seq int
	var name, path string
	if err := db.db.QueryRow("PRAGMA database_list").Scan(&seq, &name, &path); err != nil {
		t.Fatal(err)
	}
	db.close()
	reopened, err := newDatabase(filepath.Dir(path), "test-encryption-key")
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.close()
	store = selectionStore(t, reopened)
	loaded, err := store.load(context.Background(), testUserID, *result.SnapshotID)
	before, _ := json.Marshal(result)
	after, _ := json.Marshal(loaded)
	if err != nil || string(before) != string(after) {
		t.Fatalf("reopen changed manifest: %s %v", after, err)
	}
	loaded.AssetIDs[0] = "mutated caller copy"
	again, err := store.load(context.Background(), testUserID, *result.SnapshotID)
	if err != nil || again.AssetIDs[0] != selectionA {
		t.Fatal("caller can mutate persisted selection")
	}
}

func TestAISelectionInternalConsumptionHonorsDisabledAI(t *testing.T) {
	_, store, result := selectionFixture(t)
	store.enabled = false
	if _, err := store.load(context.Background(), testUserID, *result.SnapshotID); err == nil {
		t.Fatal("internal consumer bypassed disabled AI")
	}
	if _, err := store.preview(context.Background(), selection.Input{Mode: "explicit", AssetIDs: []string{selectionA}, Scope: &selection.Scope{View: "all"}}, testUserID); err == nil {
		t.Fatal("internal preview bypassed disabled AI")
	}
	if err := store.cleanup(context.Background()); err != nil {
		t.Fatal("disabled AI blocked cleanup", err)
	}
}
