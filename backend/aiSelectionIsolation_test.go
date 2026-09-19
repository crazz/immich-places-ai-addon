package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/selection"
)

func TestAISelectionOwnerIsolationAndDeletion(t *testing.T) {
	db, store, result := selectionFixture(t)
	if err := db.createUser(context.Background(), "other", "other@example.com", "hash"); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{*result.SnapshotID, "missing"} {
		loaded, err := store.load(context.Background(), "other", id)
		if !errors.Is(err, sql.ErrNoRows) || loaded.SnapshotID != nil {
			t.Fatalf("foreign/absent read: %+v %v", loaded, err)
		}
	}
	selectionSQL(t, db, "UPDATE assets SET userID='other',originalFileName='foreign-private-name' WHERE immichID=?", selectionB)
	preview, err := store.preview(context.Background(), selection.Input{Mode: "explicit", AssetIDs: []string{selectionA, selectionB, selectionID(3)}, Scope: &selection.Scope{View: "all"}}, testUserID)
	if err != nil || preview.EligibleCount != 1 || preview.Exclusions[0].Reason != "unavailable" || preview.Exclusions[1].Reason != "unavailable" {
		t.Fatalf("foreign asset leaked: %+v %v", preview, err)
	}
	data, _ := json.Marshal(preview)
	if strings.Contains(string(data), "foreign-private") {
		t.Fatal("metadata leaked")
	}
	other, err := store.preview(context.Background(), selection.Input{Mode: "explicit", AssetIDs: []string{selectionB}, Scope: &selection.Scope{View: "all"}}, "other")
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, db, "DELETE FROM users WHERE ID=?", testUserID)
	for _, table := range []string{"ai_selection_snapshots", "ai_selection_items"} {
		var count int
		if err := db.db.QueryRow("SELECT count(*) FROM "+table+" WHERE userID=?", testUserID).Scan(&count); err != nil || count != 0 {
			t.Fatalf("account cascade: %s %d %v", table, count, err)
		}
	}
	if _, err := store.load(context.Background(), "other", *other.SnapshotID); err != nil {
		t.Fatal("account deletion removed another owner's selection", err)
	}
}
