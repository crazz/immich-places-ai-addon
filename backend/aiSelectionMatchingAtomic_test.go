package main

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"immich-places-backend/internal/ai/selection"
)

func TestAISelectionMatchingEnumerationUsesOneCatalogState(t *testing.T) {
	db := newTestDB(t)
	for i := 1; i <= 3; i++ {
		seedAsset(t, db, selectionID(i), nil, nil, "2026-09-19")
	}
	tx, err := db.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	scope := selection.Scope{View: "all", GPSFilter: "no-gps", HiddenFilter: "visible"}
	scanned := make(chan struct{})
	changed := make(chan error, 1)
	go func() {
		<-scanned
		_, err := db.db.Exec("UPDATE assets SET isHidden=1 WHERE userID=?", testUserID)
		changed <- err
	}()
	observed := 0
	result, err := selection.CollectMatching(500, func(yield func(selection.Match) error) error {
		return enumerateAISelectionMatching(context.Background(), tx, testUserID, scope, func(match selection.Match) error {
			observed++
			if observed == 1 {
				close(scanned)
				if err := <-changed; err != nil {
					return err
				}
			}
			return yield(match)
		})
	})
	if err != nil || result.MatchedCount != 3 || result.EligibleCount != 3 || result.AssetIDs[0] != selectionID(3) || result.AssetIDs[2] != selectionID(1) {
		t.Fatalf("split observation: %+v %v", result, err)
	}
	_, err = tx.Exec("INSERT INTO ai_selection_snapshots VALUES (?,'failed-query',0,'{}','digest')", testUserID)
	if err == nil {
		t.Fatal("stale read transaction promoted")
	}
	rec := httptest.NewRecorder()
	writeAISelectionFailure(rec, err)
	if rec.Code != 503 {
		t.Fatalf("contention status: %d %s", rec.Code, rec.Body.String())
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&n); err != nil || n != 0 {
		t.Fatalf("partial transaction: %d %v", n, err)
	}
	after := matchingPreview(t, db, selection.Scope{View: "all"})
	if after.MatchedCount != 0 || after.SnapshotID != nil {
		t.Fatalf("next preview missed committed state: %+v", after)
	}
}

func TestAISelectionMatchingInterruptedScanDiscardsCounts(t *testing.T) {
	for _, kind := range []string{"cancel", "deadline", "scan", "storage"} {
		t.Run(kind, func(t *testing.T) {
			db := newTestDB(t)
			for i := 1; i <= 3; i++ {
				seedAsset(t, db, selectionID(i), nil, nil, "2026-09-19")
			}
			if kind == "scan" {
				selectionSQL(t, db, "UPDATE assets SET latitude='private-invalid-coordinate' WHERE immichID=?", selectionID(1))
			}
			tx, err := db.db.BeginTx(context.Background(), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			visited := 0
			result, err := selection.CollectMatching(500, func(yield func(selection.Match) error) error {
				return enumerateAISelectionMatching(ctx, tx, testUserID, selection.Scope{View: "all", GPSFilter: "no-gps", HiddenFilter: "visible"}, func(match selection.Match) error {
					visited++
					if err := yield(match); err != nil {
						return err
					}
					if visited == 1 {
						switch kind {
						case "cancel":
							cancel()
						case "deadline":
							return context.DeadlineExceeded
						case "storage":
							return errors.New("private-storage-detail")
						}
					}
					return nil
				})
			})
			if err == nil || visited == 0 || result.MatchedCount != 0 || result.AssetIDs != nil {
				t.Fatalf("partial scan: visited=%d %+v %v", visited, result, err)
			}
			rec := httptest.NewRecorder()
			writeAISelectionFailure(rec, err)
			if rec.Code < 500 || strings.Contains(rec.Body.String(), "private-") || strings.Contains(rec.Body.String(), "matchedCount") {
				t.Fatalf("unsafe partial error: %s", rec.Body.String())
			}
		})
	}
}

func TestAISelectionMatchingFailedPublicationLeavesNothing(t *testing.T) {
	for _, kind := range []string{"item", "commit", "deadline", "catalog"} {
		t.Run(kind, func(t *testing.T) {
			db := newTestDB(t)
			store := selectionStore(t, db)
			seedAsset(t, db, selectionA, nil, nil, "2026-09-19")
			switch kind {
			case "item":
				selectionSQL(t, db, `CREATE TRIGGER reject_query_item BEFORE INSERT ON ai_selection_items BEGIN SELECT RAISE(ABORT,'private-item-failure'); END`)
			case "commit":
				selectionSQL(t, db, `CREATE TABLE query_commit_failure (owner TEXT REFERENCES users(ID) DEFERRABLE INITIALLY DEFERRED)`)
				selectionSQL(t, db, `CREATE TRIGGER reject_query_commit AFTER INSERT ON ai_selection_snapshots BEGIN INSERT INTO query_commit_failure VALUES ('missing'); END`)
			case "catalog":
				selectionSQL(t, db, "ALTER TABLE assets RENAME TO unavailable_assets")
			}
			ctx := context.Background()
			if kind == "deadline" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, time.Now().Add(-time.Second))
				defer cancel()
			}
			result, err := store.preview(ctx, matchingInput(selection.Scope{View: "all"}), testUserID)
			if err == nil || result.SnapshotID != nil || result.QuerySummary != nil {
				t.Fatalf("partial publication: %+v %v", result, err)
			}
			for _, table := range []string{"ai_selection_snapshots", "ai_selection_items"} {
				var n int
				if err := db.db.QueryRow("SELECT count(*) FROM " + table).Scan(&n); err != nil || n != 0 {
					t.Fatalf("partial %s: %d %v", table, n, err)
				}
			}
		})
	}
}
