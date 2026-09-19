package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/selection"
)

func TestAISelectionUsesCaptureTimestampWithoutFileDateFallback(t *testing.T) {
	db := newTestDB(t)
	store := selectionStore(t, db)
	ids := []string{selectionA, selectionB, selectionID(3)}
	seedAsset(t, db, ids[0], nil, nil, "2026-09-19")
	seedAsset(t, db, ids[1], nil, nil, "2026-01-01")
	seedAsset(t, db, ids[2], nil, nil, "2026-01-01")
	selectionSQL(t, db, "UPDATE assets SET dateTimeOriginal='2026-01-01T23:30:00-12:00' WHERE immichID=?", ids[0])
	selectionSQL(t, db, "UPDATE assets SET dateTimeOriginal=NULL WHERE immichID=?", ids[1])
	selectionSQL(t, db, "UPDATE assets SET dateTimeOriginal='2026-09-19T00:30:00+14:00' WHERE immichID=?", ids[2])
	input := selection.Input{Mode: "explicit", AssetIDs: ids, Scope: &selection.Scope{View: "all", StartDate: "2026-01-01", EndDate: "2026-01-01"}}
	result, err := store.preview(context.Background(), input, testUserID)
	if err != nil || result.EligibleCount != 1 || result.AssetIDs[0] != ids[0] {
		t.Fatalf("selection disagrees with capture-date contract: %+v %v", result, err)
	}
	selectionSQL(t, db, "UPDATE assets SET fileCreatedAt='2025-12-31' WHERE immichID=?", ids[0])
	if _, err := store.load(context.Background(), testUserID, *result.SnapshotID); err != nil {
		t.Fatalf("unrelated file date staled capture scope: %v", err)
	}
	selectionSQL(t, db, "UPDATE assets SET dateTimeOriginal='2026-01-02' WHERE immichID=?", ids[0])
	if _, err := store.load(context.Background(), testUserID, *result.SnapshotID); err == nil {
		t.Fatal("changed capture date did not stale snapshot")
	}
}
