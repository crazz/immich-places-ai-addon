package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"immich-places-backend/internal/ai/selection"
)

func TestAISelectionMatchingReopenPreservesHistoricalMembership(t *testing.T) {
	db := newTestDB(t)
	for i := 1; i <= 3; i++ {
		seedAsset(t, db, selectionID(i), nil, nil, "2026-09-19")
	}
	selectionSQL(t, db, "UPDATE assets SET type='VIDEO' WHERE immichID=?", selectionID(3))
	original := matchingPreview(t, db, selection.Scope{View: "all"})
	seedAsset(t, db, selectionID(4), nil, nil, "2026-10-01")
	selectionSQL(t, db, "UPDATE assets SET type='IMAGE',fileCreatedAt='2026-10-02',syncedAt='later' WHERE immichID=?", selectionID(3))
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
	store := selectionStore(t, reopened)
	loaded, err := store.load(context.Background(), testUserID, *original.SnapshotID)
	before, _ := json.Marshal(original)
	after, _ := json.Marshal(loaded)
	if err != nil || string(before) != string(after) || loaded.MatchedCount != 3 || loaded.ExcludedCount != 1 {
		t.Fatalf("query reran on reopen: %s %v", after, err)
	}
	loaded.ExclusionCounts["unsupported_type"] = 999
	loaded.AssetIDs[0] = "caller mutation"
	again, err := store.load(context.Background(), testUserID, *original.SnapshotID)
	if err != nil || again.ExclusionCounts["unsupported_type"] != 1 || again.AssetIDs[0] != selectionID(2) {
		t.Fatalf("mutable persisted result: %+v %v", again, err)
	}
	selectionSQL(t, reopened, "UPDATE assets SET isHidden=1 WHERE immichID=?", selectionID(1))
	stale, err := store.load(context.Background(), testUserID, *original.SnapshotID)
	if !errors.Is(err, errAISelectionStale) || stale.SnapshotID != nil {
		t.Fatalf("stale result retained subset: %+v %v", stale, err)
	}
}

func TestAISelectionMatchingSharedQuotaExpiryAndEpoch(t *testing.T) {
	db, store, explicit := selectionFixture(t)
	input := matchingInput(selection.Scope{View: "all"})
	for i := 0; i < 19; i++ {
		if _, err := store.preview(context.Background(), input, testUserID); err != nil {
			t.Fatal(err)
		}
	}
	for _, mode := range []string{"explicit", "all-matching"} {
		attempt := input
		if mode == "explicit" {
			attempt = selection.Input{Mode: mode, AssetIDs: []string{selectionA}, Scope: input.Scope}
		}
		if result, err := store.preview(context.Background(), attempt, testUserID); !errors.Is(err, errAISelectionOwnerQuota) || result.SnapshotID != nil {
			t.Fatalf("mixed-mode quota bypass: %s %v", mode, err)
		}
	}
	store.now = func() time.Time { return explicit.ExpiresAt.Add(time.Hour) }
	renewed, err := store.preview(context.Background(), input, testUserID)
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&n); err != nil || n != 1 {
		t.Fatalf("shared cleanup: %d %v", n, err)
	}
	store.now = func() time.Time { return renewed.ExpiresAt }
	if _, err := store.load(context.Background(), testUserID, *renewed.SnapshotID); !errors.Is(err, errAISelectionStale) {
		t.Fatalf("query expiry ignored: %v", err)
	}
	if err := store.bind(context.Background(), "https://immich.example/api", "2"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.load(context.Background(), testUserID, *renewed.SnapshotID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("epoch retained query: %v", err)
	}
}

func TestAISelectionMatchingAuthorityAndAccountDeletion(t *testing.T) {
	db, _ := aiTestHandler(t, true)
	seedAsset(t, db, selectionA, nil, nil, "2026-09-19")
	cfg := &Config{AIEnabled: true, ImmichURL: "https://immich.example/api", AIPublicOrigin: aiTestOrigin}
	h := newAISelectionHandler(db, cfg)
	body := `{"mode":"all-matching","scope":{"view":"all"}}`
	for _, tc := range []struct {
		session bool
		origin  string
		enabled bool
		status  int
	}{
		{false, aiTestOrigin, true, 401}, {true, "https://foreign.example", true, 403}, {true, "", true, 403}, {true, aiTestOrigin, false, 503},
	} {
		cfg.AIEnabled = tc.enabled
		rec := aiRequest(newAISelectionHandler(db, cfg), "POST", "/ai/selection-preview", body, tc.origin, tc.session)
		if rec.Code != tc.status {
			t.Fatalf("query authority: %d %s", rec.Code, rec.Body.String())
		}
	}
	cfg.AIEnabled = true
	if err := db.createUser(context.Background(), "other", "other@example.com", "hash"); err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, db, "INSERT INTO assets (userID,immichID,type,originalFileName,fileCreatedAt) VALUES ('other',?,'IMAGE','private.jpg','2026-09-19')", selectionB)
	foreign, err := h.store.preview(context.Background(), matchingInput(selection.Scope{View: "all"}), "other")
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{*foreign.SnapshotID, "absent"} {
		rec := aiRequest(h, "GET", "/ai/selections/"+id, "", "", true)
		if rec.Code != 404 {
			t.Fatalf("foreign resource response: %d %s", rec.Code, rec.Body.String())
		}
	}
	mine, err := h.store.preview(context.Background(), matchingInput(selection.Scope{View: "all"}), testUserID)
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, db, "DELETE FROM users WHERE ID=?", testUserID)
	for _, table := range []string{"ai_selection_snapshots", "ai_selection_items"} {
		var n int
		if err := db.db.QueryRow("SELECT count(*) FROM "+table+" WHERE userID=?", testUserID).Scan(&n); err != nil || n != 0 {
			t.Fatalf("query account cleanup: %s %d %v", table, n, err)
		}
	}
	if _, err := h.store.load(context.Background(), testUserID, *mine.SnapshotID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("deleted resource survived", err)
	}
	if _, err := h.store.load(context.Background(), "other", *foreign.SnapshotID); err != nil {
		t.Fatal("other owner affected", err)
	}
}
