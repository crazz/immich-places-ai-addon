package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"immich-places-backend/internal/ai/selection"
)

func TestAISelectionMatchingHundredThousandCatalog(t *testing.T) {
	db := newTestDB(t)
	store := selectionStore(t, db)
	selectionSQL(t, db, `WITH RECURSIVE n(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM n WHERE x<100000)
 INSERT INTO assets (userID,immichID,type,originalFileName,fileCreatedAt)
 SELECT ?,printf('aaaaaaaa-0000-4000-8000-%012d',x),'VIDEO','synthetic.jpg','2026-09-19' FROM n`, testUserID)
	retained := 0
	for _, eligible := range []int{10, 500, 100000} {
		selectionSQL(t, db, "UPDATE assets SET type='IMAGE' WHERE immichID<=?", selectionID(eligible))
		var before, after runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&before)
		start := time.Now()
		result, err := store.preview(context.Background(), matchingInput(selection.Scope{View: "all"}), testUserID)
		elapsed := time.Since(start)
		runtime.ReadMemStats(&after)
		var limit *selection.MatchingLimitError
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			// Instrumented/slow machines may exceed the real production deadline. That
			// outcome must fail closed, never masquerade as complete catalog evidence.
			rec := httptest.NewRecorder()
			writeAISelectionFailure(rec, err)
			var wire map[string]json.RawMessage
			if decodeErr := json.Unmarshal(rec.Body.Bytes(), &wire); decodeErr != nil {
				t.Fatal(decodeErr)
			}
			if result.SnapshotID != nil || result.QuerySummary != nil || len(result.AssetIDs) != 0 || rec.Code != 503 || wire["matchedCount"] != nil {
				t.Fatalf("partial deadline result: %+v %s", result, rec.Body.String())
			}
			t.Logf("100000 matches / %d eligible: DEADLINE, zero publication; elapsed=%s totalAllocated=%d bytes", eligible, elapsed, after.TotalAlloc-before.TotalAlloc)
		case eligible > 500:
			if !errors.As(err, &limit) || limit.MatchedCount != 100000 || limit.EligibleCount != eligible || result.SnapshotID != nil {
				t.Fatalf("large overflow: %v", err)
			}
			t.Logf("100000 eligible: exact overflow; elapsed=%s totalAllocated=%d bytes", elapsed, after.TotalAlloc-before.TotalAlloc)
		default:
			if err != nil || result.MatchedCount != 100000 || result.EligibleCount != eligible || result.ExcludedCount != 100000-eligible || len(result.AssetIDs) != eligible {
				t.Fatalf("large matching catalog: %+v %v", result, err)
			}
			retained++
			t.Logf("100000 matches / %d eligible: complete; elapsed=%s totalAllocated=%d bytes", eligible, elapsed, after.TotalAlloc-before.TotalAlloc)
		}
		var count int
		if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&count); err != nil || count != retained {
			t.Fatalf("partial or missing publication: %d want %d err=%v", count, retained, err)
		}
	}
	t.Log("Synthetic local measurements; not NAS latency or hardware acceptance.")
}
