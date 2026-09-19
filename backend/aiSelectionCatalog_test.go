package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"immich-places-backend/internal/ai/selection"
	"testing"
)

func selectionID(n int) string { return fmt.Sprintf("aaaaaaaa-0000-4000-8000-%012d", n) }
func selectionSQL(t *testing.T, db *Database, query string, args ...any) {
	t.Helper()
	if _, err := db.db.Exec(query, args...); err != nil {
		t.Fatal(err)
	}
}

func TestAISelectionCatalogExclusions(t *testing.T) {
	db, _ := aiTestHandler(t, true)
	ids := []string{}
	for n := 1; n <= 7; n++ {
		id := selectionID(n)
		ids = append(ids, id)
		seedAsset(t, db, id, nil, nil, "2026-09-19T12:00:00")
	}
	selectionSQL(t, db, "UPDATE assets SET type='VIDEO',isHidden=1 WHERE immichID=?", ids[1])
	selectionSQL(t, db, "UPDATE assets SET isHidden=1 WHERE immichID=?", ids[2])
	selectionSQL(t, db, "UPDATE assets SET stackPrimaryAssetID=? WHERE immichID=?", ids[0], ids[3])
	selectionSQL(t, db, "UPDATE assets SET latitude=0,longitude=0 WHERE immichID=?", ids[4])
	selectionSQL(t, db, "INSERT INTO libraries (libraryID,isHidden) VALUES ('private-library',1)")
	selectionSQL(t, db, "UPDATE assets SET libraryID='private-library' WHERE immichID=?", ids[5])
	selectionSQL(t, db, "DELETE FROM assets WHERE immichID=?", ids[6])
	body, _ := json.Marshal(map[string]any{"mode": "explicit", "assetIDs": ids, "scope": map[string]string{"view": "all", "hiddenFilter": "all"}})
	h := newAISelectionHandler(db, &Config{AIEnabled: true, ImmichURL: "https://immich.example/api", AIPublicOrigin: aiTestOrigin})
	rec := aiRequest(h, "POST", "/ai/selection-preview", string(body), aiTestOrigin, true)
	var manifest struct {
		AssetIDs                     []string
		EligibleCount, ExcludedCount int
		Exclusions                   []struct{ AssetID, Reason string }
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &manifest); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 || manifest.EligibleCount != 1 || manifest.ExcludedCount != 6 || len(manifest.AssetIDs) != 1 || manifest.AssetIDs[0] != ids[0] {
		t.Fatalf("bad selection: %s", rec.Body.String())
	}
	want := []string{"unsupported_type", "hidden_by_policy", "stack_child", "outside_scope", "unavailable", "unavailable"}
	for i, e := range manifest.Exclusions {
		if e.AssetID != ids[i+1] || e.Reason != want[i] {
			t.Fatalf("exclusion: %+v", e)
		}
	}
}

func TestAISelectionZeroEligibleDoesNotPersist(t *testing.T) {
	db, _ := aiTestHandler(t, true)
	h := newAISelectionHandler(db, &Config{AIEnabled: true, ImmichURL: "https://immich.example/api", AIPublicOrigin: aiTestOrigin})
	rec := aiRequest(h, "POST", "/ai/selection-preview", selectionBody, aiTestOrigin, true)
	var result struct {
		SnapshotID                                   *string
		EligibleCount, ExcludedCount, RequestedCount int
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 || result.SnapshotID != nil || result.EligibleCount != 0 || result.ExcludedCount != 2 || result.RequestedCount != 3 || n != 0 {
		t.Fatalf("empty snapshot: %s count=%d", rec.Body.String(), n)
	}
}

func TestAISelectionRecursiveFolderScope(t *testing.T) {
	db := newTestDB(t)
	ids := []string{}
	for i, path := range []string{"/Trip/a.jpg", "/Trip/Day1/b.jpg", "/Trips/c.jpg", "/trip/d.jpg"} {
		id := selectionID(i + 1)
		ids = append(ids, id)
		seedAsset(t, db, id, nil, nil, "2026-09-19")
		selectionSQL(t, db, "UPDATE assets SET originalPath=? WHERE immichID=?", path, id)
	}
	result, err := selectionStore(t, db).preview(context.Background(), selection.Input{Mode: "explicit", AssetIDs: ids, Scope: &selection.Scope{View: "folder", FolderPath: "/Trip/"}}, testUserID)
	if err != nil || result.EligibleCount != 2 || result.ExcludedCount != 2 || result.Scope.FolderPath != "/Trip" {
		t.Fatalf("folder selection: %+v %v", result, err)
	}
	if result.AssetIDs[0] != ids[0] || result.AssetIDs[1] != ids[1] {
		t.Fatal(result.AssetIDs)
	}
}

func TestAISelectionCaptureDatesAndGPS(t *testing.T) {
	db := newTestDB(t)
	ids := []string{}
	for i, date := range []string{"2026-09-19T23:59:00-12:00", "2026-09-19T00:01:00+14:00", "2026-09-19T12:00:00", "", "2026-09-18T12:00:00", "2026-02-30T12:00:00"} {
		id := selectionID(i + 1)
		ids = append(ids, id)
		seedAsset(t, db, id, nil, nil, "2026-10-01")
		selectionSQL(t, db, "UPDATE assets SET dateTimeOriginal=? WHERE immichID=?", date, id)
	}
	selectionSQL(t, db, "UPDATE assets SET latitude=0 WHERE immichID=?", ids[0])
	selectionSQL(t, db, "UPDATE assets SET latitude=0,longitude=0 WHERE immichID=?", ids[2])
	for _, tc := range []struct {
		start, end, gps string
		eligible        int
	}{
		{"2026-09-19", "2026-09-19", "no-gps", 2}, {"2026-09-19", "", "with-gps", 1}, {"", "2026-09-19", "all", 4},
	} {
		result, err := selectionStore(t, db).preview(context.Background(), selection.Input{Mode: "explicit", AssetIDs: ids, Scope: &selection.Scope{View: "all", GPSFilter: tc.gps, StartDate: tc.start, EndDate: tc.end}}, testUserID)
		if err != nil || result.EligibleCount != tc.eligible {
			t.Fatalf("date scope %+v: %+v %v", tc, result, err)
		}
	}
}

func TestAISelectionAlbumTagScope(t *testing.T) {
	db := newTestDB(t)
	for _, id := range []string{selectionA, selectionB} {
		seedAsset(t, db, id, nil, nil, "2026-09-19")
	}
	selectionSQL(t, db, "INSERT INTO albums (userID,immichID,albumName,updatedAt) VALUES (?,'album','private album','')", testUserID)
	selectionSQL(t, db, "INSERT INTO tags (userID,immichID,name,value,updatedAt) VALUES (?,'tag','private tag','tag','')", testUserID)
	selectionSQL(t, db, "INSERT INTO albumAssets VALUES (?,'album',?)", testUserID, selectionA)
	selectionSQL(t, db, "INSERT INTO assetTags VALUES (?,'tag',?)", testUserID, selectionA)
	input := selection.Input{Mode: "explicit", AssetIDs: []string{selectionA, selectionB}, Scope: &selection.Scope{View: "album", AlbumID: "album", TagID: "tag"}}
	result, err := selectionStore(t, db).preview(context.Background(), input, testUserID)
	if err != nil || result.EligibleCount != 1 || result.AssetIDs[0] != selectionA || result.Exclusions[0].Reason != "outside_scope" {
		t.Fatalf("album/tag: %+v %v", result, err)
	}
	for _, scope := range []selection.Scope{{View: "album", AlbumID: "missing"}, {View: "all", TagID: "missing"}} {
		input.Scope = &scope
		_, err = selectionStore(t, db).preview(context.Background(), input, testUserID)
		if !errors.Is(err, selection.ErrInvalid) {
			t.Fatalf("missing scope accepted: %v", err)
		}
	}
}

func TestAISelectionHiddenOnlyScopeRetainsPolicyAndNoResource(t *testing.T) {
	db := newTestDB(t)
	store := selectionStore(t, db)
	seedAsset(t, db, selectionA, nil, nil, "2026-09-19")
	selectionSQL(t, db, "UPDATE assets SET isHidden=1 WHERE immichID=?", selectionA)
	result, err := store.preview(context.Background(), selection.Input{Mode: "explicit", AssetIDs: []string{selectionA}, Scope: &selection.Scope{View: "all", HiddenFilter: "hidden"}}, testUserID)
	if err != nil || result.SnapshotID != nil || result.EligibleCount != 0 || result.ExcludedCount != 1 || result.Scope.HiddenFilter != "hidden" || result.Exclusions[0].Reason != "hidden_by_policy" {
		t.Fatalf("hidden scope: %+v %v", result, err)
	}
}
