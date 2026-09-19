package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"immich-places-backend/internal/ai/selection"
)

func TestAISelectionOwnerQuotaPreservesLiveSnapshots(t *testing.T) {
	db, store, first := selectionFixture(t)
	input := selection.Input{Mode: "explicit", AssetIDs: []string{selectionA}, Scope: &selection.Scope{View: "all"}}
	for i := 1; i < 20; i++ {
		if _, err := store.preview(context.Background(), input, testUserID); err != nil {
			t.Fatal(err)
		}
	}
	if result, err := store.preview(context.Background(), input, testUserID); err == nil || result.SnapshotID != nil {
		t.Fatal("owner quota did not reject whole request")
	}
	if _, err := store.load(context.Background(), testUserID, *first.SnapshotID); err != nil {
		t.Fatalf("quota evicted live selection: %v", err)
	}
	var count int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&count); err != nil || count != 20 {
		t.Fatalf("quota count: %d %v", count, err)
	}
}

func TestAISelectionGlobalCapacityCountsRetainedHeaders(t *testing.T) {
	db, store, first := selectionFixture(t)
	if err := db.createUser(context.Background(), "other", "other@example.com", "hash"); err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, db, `WITH RECURSIVE n(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM n WHERE x<999)
 INSERT INTO ai_selection_snapshots (userID,id,expiresAt,manifest,digest) SELECT 'other',CAST(x AS TEXT),?,'{}','unused' FROM n`, first.ExpiresAt.UnixNano())
	result, err := store.preview(context.Background(), selection.Input{Mode: "explicit", AssetIDs: []string{selectionA}, Scope: &selection.Scope{View: "all"}}, testUserID)
	if err == nil || result.SnapshotID != nil {
		t.Fatal("global capacity bypassed")
	}
}

func TestAISelectionCleanupIsBoundedAndCascades(t *testing.T) {
	db, store, first := selectionFixture(t)
	selectionSQL(t, db, `WITH RECURSIVE n(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM n WHERE x<205)
 INSERT INTO ai_selection_snapshots (userID,id,expiresAt,manifest,digest) SELECT ?,CAST(x AS TEXT),0,'{}','unused' FROM n`, testUserID)
	selectionSQL(t, db, `INSERT INTO ai_selection_items (userID,snapshotID,position,assetID,facts)
 SELECT userID,id,0,'not-a-catalog-row','{}' FROM ai_selection_snapshots WHERE expiresAt=0`)
	for _, want := range []int{106, 6, 1} {
		if err := store.cleanup(context.Background()); err != nil {
			t.Fatal(err)
		}
		var headers, items int
		if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&headers); err != nil {
			t.Fatal(err)
		}
		if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_items").Scan(&items); err != nil {
			t.Fatal(err)
		}
		if headers != want || items != want+1 {
			t.Fatalf("cleanup unbounded or incomplete: headers=%d items=%d want=%d", headers, items, want)
		}
	}
	if _, err := store.load(context.Background(), testUserID, *first.SnapshotID); err != nil {
		t.Fatal(err)
	}
}

func TestAISelectionCreationCleansExpiredBeforeQuota(t *testing.T) {
	db, store, first := selectionFixture(t)
	store.now = func() time.Time { return first.ExpiresAt }
	result, err := store.preview(context.Background(), selection.Input{Mode: "explicit", AssetIDs: []string{selectionA}, Scope: &selection.Scope{View: "all"}}, testUserID)
	if err != nil || result.SnapshotID == nil {
		t.Fatalf("preview: %+v %v", result, err)
	}
	var count int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&count); err != nil || count != 1 {
		t.Fatalf("expired retained at creation: %d %v", count, err)
	}
}

func TestAISelectionResponseLimitRejectsWholePreview(t *testing.T) {
	db := newTestDB(t)
	store := selectionStore(t, db)
	ids := make([]string, 500)
	for i := range ids {
		ids[i] = selectionID(i + 1)
	}
	input := selection.Input{Mode: "explicit", AssetIDs: ids, Scope: &selection.Scope{View: "folder", FolderPath: "/" + strings.Repeat("a", selection.MaxBytes-23000)}}
	data, err := json.Marshal(input)
	if err != nil || len(data) > selection.MaxBytes {
		t.Fatalf("invalid fixture: %d %v", len(data), err)
	}
	result, err := store.preview(context.Background(), input, testUserID)
	if !errors.Is(err, selection.ErrLimit) || result.SnapshotID != nil {
		t.Fatalf("oversized response accepted: %v", err)
	}
}

func TestAISelectionDisabledStartupCleansExpired(t *testing.T) {
	db, store, first := selectionFixture(t)
	store.now = func() time.Time { return first.ExpiresAt }
	selectionSQL(t, db, "UPDATE ai_selection_snapshots SET expiresAt=0")
	h := newAISelectionHandler(db, &Config{ImmichURL: "https://immich.example/api", AIEnabled: false})
	if rec := aiRequest(h, "GET", "/ai/selections/"+*first.SnapshotID, "", "", false); rec.Code != 401 {
		t.Fatal(rec.Code)
	}
	var count int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&count); err != nil || count != 0 {
		t.Fatalf("AI-disabled startup retained expiry: %d %v", count, err)
	}
}

func TestAISelectionPeriodicCleanupFailureRetriesAndCancels(t *testing.T) {
	db, store, first := selectionFixture(t)
	store.now = func() time.Time { return first.ExpiresAt }
	selectionSQL(t, db, `CREATE TRIGGER reject_cleanup BEFORE DELETE ON ai_selection_snapshots BEGIN SELECT RAISE(ABORT,'private path and IDs'); END`)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ticks := make(chan time.Time)
	failures := make(chan struct{}, 1)
	done := make(chan struct{})
	go func() { store.runCleanup(ctx, ticks, func() { failures <- struct{}{} }); close(done) }()
	select {
	case ticks <- store.now():
	case <-time.After(time.Second):
		t.Fatal("cleanup loop not running")
	}
	select {
	case <-failures:
	case <-time.After(time.Second):
		t.Fatal("cleanup failure not observable")
	}
	if _, err := store.load(context.Background(), testUserID, *first.SnapshotID); err == nil {
		t.Fatal("cleanup failure made expiry readable")
	}
	selectionSQL(t, db, "DROP TRIGGER reject_cleanup")
	select {
	case ticks <- store.now():
	case <-time.After(time.Second):
		t.Fatal("cleanup did not retry")
	}
	// The next receive begins only after the preceding cleanup completed.
	select {
	case ticks <- store.now():
	case <-time.After(time.Second):
		t.Fatal("cleanup did not finish")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("cleanup ignored cancellation")
	}
	var count int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&count); err != nil || count != 0 {
		t.Fatalf("cleanup retry failed: %d %v", count, err)
	}
}
