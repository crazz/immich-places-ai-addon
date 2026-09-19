package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"immich-places-backend/internal/ai/selection"
)

func matchingInput(scope selection.Scope) selection.Input {
	return selection.Input{Mode: "all-matching", Scope: &scope}
}

func TestAISelectionMatchingWholeCatalogOrderedRoundTrip(t *testing.T) {
	db, _ := aiTestHandler(t, true)
	for i := 1; i <= 205; i++ {
		seedAsset(t, db, selectionID(i), nil, nil, "2026-09-19")
	}
	handler := newAISelectionHandler(db, &Config{AIEnabled: true, ImmichURL: "https://immich.example/api", AIPublicOrigin: aiTestOrigin})
	rec := aiRequest(handler, "POST", "/ai/selection-preview", `{"mode":"all-matching","scope":{"view":"all"}}`, aiTestOrigin, true)
	var result selection.Manifest
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 || result.SnapshotID == nil || result.EligibleCount != 205 || len(result.AssetIDs) != 205 || result.RequestedCount != 205 || result.UniqueCount != 205 || result.DuplicateCount != 0 {
		t.Fatalf("whole catalog: %d %s", rec.Code, rec.Body.String())
	}
	for i, id := range result.AssetIDs {
		if id != selectionID(205-i) {
			t.Fatalf("unstable ordering at %d: %s", i, id)
		}
	}
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &wire); err != nil {
		t.Fatal(err)
	}
	if string(wire["matchedCount"]) != "205" || string(wire["exclusionCounts"]) != "{}" || wire["exclusions"] != nil {
		t.Fatalf("query response shape: %s", rec.Body.String())
	}
	read := aiRequest(handler, "GET", "/ai/selections/"+*result.SnapshotID, "", "", true)
	if read.Code != 200 || read.Body.String() != rec.Body.String() {
		t.Fatalf("frozen roundtrip: %d %s", read.Code, read.Body.String())
	}
}

func matchingPreview(t *testing.T, db *Database, scope selection.Scope) selection.Manifest {
	t.Helper()
	result, err := selectionStore(t, db).preview(context.Background(), matchingInput(scope), testUserID)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func matchingIDs(t *testing.T, result selection.Manifest, expected ...int) {
	t.Helper()
	ids := make([]string, len(expected))
	for i, n := range expected {
		ids[i] = selectionID(n)
	}
	if fmt.Sprint(result.AssetIDs) != fmt.Sprint(ids) {
		t.Fatalf("membership got %v want %v", result.AssetIDs, ids)
	}
}

func TestAISelectionMatchingCountsEmptyAndMixed(t *testing.T) {
	db := newTestDB(t)
	empty := matchingPreview(t, db, selection.Scope{View: "all"})
	if empty.SnapshotID != nil || empty.QuerySummary == nil || empty.MatchedCount != 0 || empty.ExclusionCounts == nil {
		t.Fatalf("empty summary: %+v", empty)
	}
	for i := 1; i <= 4; i++ {
		seedAsset(t, db, selectionID(i), nil, nil, "2026-09-19")
	}
	selectionSQL(t, db, "UPDATE assets SET type='VIDEO' WHERE immichID=?", selectionID(3))
	selectionSQL(t, db, "UPDATE assets SET isHidden=1 WHERE immichID=?", selectionID(4))
	mixed := matchingPreview(t, db, selection.Scope{View: "all", HiddenFilter: "all"})
	matchingIDs(t, mixed, 2, 1)
	if mixed.MatchedCount != 4 || mixed.EligibleCount != 2 || mixed.ExcludedCount != 2 || mixed.ExclusionCounts["unsupported_type"] != 1 || mixed.ExclusionCounts["hidden_by_policy"] != 1 {
		t.Fatalf("mixed summary: %+v", mixed)
	}
	hidden := matchingPreview(t, db, selection.Scope{View: "all", HiddenFilter: "hidden"})
	if hidden.SnapshotID != nil || hidden.MatchedCount != 1 || hidden.EligibleCount != 0 || hidden.ExcludedCount != 1 {
		t.Fatalf("excluded only: %+v", hidden)
	}
	var count int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&count); err != nil || count != 1 {
		t.Fatalf("empty resource retained: %d %v", count, err)
	}
}
