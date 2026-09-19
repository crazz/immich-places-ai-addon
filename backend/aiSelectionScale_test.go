package main

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"immich-places-backend/internal/ai/selection"
)

func TestAISelectionSyntheticHundredThousandCatalog(t *testing.T) {
	db := newTestDB(t)
	store := selectionStore(t, db)
	selectionSQL(t, db, `WITH RECURSIVE n(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM n WHERE x<100000)
 INSERT INTO assets (userID,immichID,type,originalFileName,fileCreatedAt)
 SELECT ?,printf('aaaaaaaa-0000-4000-8000-%012d',x),'IMAGE','synthetic.jpg','2026-09-19' FROM n`, testUserID)
	ids := make([]string, 500)
	for i := range ids {
		ids[i] = selectionID(i + 1)
	}
	input := selection.Input{Mode: "explicit", AssetIDs: ids, Scope: &selection.Scope{View: "all"}}
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	start := time.Now()
	result, err := store.preview(context.Background(), input, testUserID)
	elapsed := time.Since(start)
	runtime.ReadMemStats(&after)
	if err != nil || result.EligibleCount != 500 || len(result.AssetIDs) != 500 {
		t.Fatalf("scale preview: %d %v", result.EligibleCount, err)
	}
	t.Logf("synthetic catalog=100000 selected=500 elapsed=%s allocated=%d bytes (local host, not NAS)", elapsed, after.TotalAlloc-before.TotalAlloc)
	input.AssetIDs = append(ids, selectionID(501))
	if rejected, err := store.preview(context.Background(), input, testUserID); !errors.Is(err, selection.ErrLimit) || rejected.SnapshotID != nil {
		t.Fatalf("configured limit did not reject whole batch: %v", err)
	}
	var count int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&count); err != nil || count != 1 {
		t.Fatalf("partial overflow resource: %d %v", count, err)
	}
}
