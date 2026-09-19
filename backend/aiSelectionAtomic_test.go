package main

import (
	"context"
	"net/http/httptest"
	"testing"

	"immich-places-backend/internal/ai/selection"
)

func TestAISelectionPublicationFailureLeavesNoPartialSnapshot(t *testing.T) {
	for _, failure := range []string{"item", "commit", "cancel"} {
		t.Run(failure, func(t *testing.T) {
			db := newTestDB(t)
			store := selectionStore(t, db)
			seedAsset(t, db, selectionA, nil, nil, "2026-09-19")
			switch failure {
			case "item":
				selectionSQL(t, db, `CREATE TRIGGER reject_item BEFORE INSERT ON ai_selection_items BEGIN SELECT RAISE(ABORT,'private failure detail'); END`)
			case "commit":
				selectionSQL(t, db, `CREATE TABLE commit_failure (owner TEXT REFERENCES users(ID) DEFERRABLE INITIALLY DEFERRED)`)
				selectionSQL(t, db, `CREATE TRIGGER reject_commit AFTER INSERT ON ai_selection_snapshots BEGIN INSERT INTO commit_failure VALUES ('not-an-owner'); END`)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if failure == "cancel" {
				cancel()
			}
			result, err := store.preview(ctx, selection.Input{Mode: "explicit", AssetIDs: []string{selectionA}, Scope: &selection.Scope{View: "all"}}, testUserID)
			if err == nil || result.SnapshotID != nil {
				t.Fatal("partial publication returned")
			}
			for _, table := range []string{"ai_selection_snapshots", "ai_selection_items"} {
				var count int
				if err := db.db.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 0 {
					t.Fatalf("rollback: %s %d %v", table, count, err)
				}
			}
		})
	}
}

func TestAISelectionConcurrentCatalogObservationFailsAtomicallyOnPromotion(t *testing.T) {
	db, _, _ := selectionFixture(t)
	ctx := context.Background()
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	scope := selection.Scope{View: "all", GPSFilter: "no-gps", HiddenFilter: "visible"}
	first, err := resolveAISelectionCandidate(ctx, tx, testUserID, selectionA, scope)
	if err != nil || first.Hidden {
		t.Fatalf("first observation: %+v %v", first, err)
	}
	written := make(chan error, 1)
	go func() {
		_, err := db.db.Exec("UPDATE assets SET isHidden=1 WHERE userID=?", testUserID)
		written <- err
	}()
	if err := <-written; err != nil {
		t.Fatal(err)
	}
	second, err := resolveAISelectionCandidate(ctx, tx, testUserID, selectionB, scope)
	if err != nil || second.Hidden {
		t.Fatalf("mixed catalog observations: %+v %v", second, err)
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO ai_selection_snapshots VALUES (?,'failed-publication',0,'{}','digest')", testUserID)
	if err == nil {
		t.Fatal("stale transaction published after competing sync")
	}
	rec := httptest.NewRecorder()
	writeAISelectionFailure(rec, err)
	if rec.Code != 503 {
		t.Fatalf("busy snapshot should be retryable 503: %d %s", rec.Code, rec.Body.String())
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots WHERE id='failed-publication'").Scan(&count); err != nil || count != 0 {
		t.Fatalf("partial publication: %d %v", count, err)
	}
}

func TestAISelectionConcurrentCreatorsCannotExceedOwnerQuota(t *testing.T) {
	db, store, _ := selectionFixture(t)
	input := selection.Input{Mode: "explicit", AssetIDs: []string{selectionA}, Scope: &selection.Scope{View: "all"}}
	for i := 1; i < 18; i++ {
		if _, err := store.preview(context.Background(), input, testUserID); err != nil {
			t.Fatal(err)
		}
	}
	start := make(chan struct{})
	results := make(chan error, 8)
	for i := 0; i < 8; i++ {
		go func() { <-start; _, err := store.preview(context.Background(), input, testUserID); results <- err }()
	}
	close(start)
	for i := 0; i < 8; i++ {
		if err := <-results; err != nil {
			rec := httptest.NewRecorder()
			writeAISelectionFailure(rec, err)
			if rec.Code != 429 && rec.Code != 503 {
				t.Fatalf("unexpected competing-creator failure: %s", rec.Body.String())
			}
		}
	}
	var count int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots WHERE userID=?", testUserID).Scan(&count); err != nil || count > 20 || count < 18 {
		t.Fatalf("concurrent quota: %d %v", count, err)
	}
}
