package main

import (
	"context"
	"immich-places-backend/internal/ai/selection"
	"testing"
)

func TestAISelectionMatchingScopeParityAndSuppression(t *testing.T) {
	db := newTestDB(t)
	if err := db.createUser(context.Background(), "other", "other@example.com", "hash"); err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, db, "INSERT INTO libraries (libraryID,isHidden) VALUES ('hidden-library',1)")
	for _, id := range []string{"album", "second-album"} {
		selectionSQL(t, db, "INSERT INTO albums (userID,immichID,albumName,updatedAt) VALUES (?,?,?,'')", testUserID, id, id)
	}
	for _, id := range []string{"tag", "second-tag"} {
		selectionSQL(t, db, "INSERT INTO tags (userID,immichID,name,value,updatedAt) VALUES (?,?,?,?,'')", testUserID, id, id, id)
	}
	for i := 1; i <= 12; i++ {
		seedAsset(t, db, selectionID(i), nil, nil, "2026-10-01")
		selectionSQL(t, db, "UPDATE assets SET originalPath='/Trip/a.jpg',dateTimeOriginal='2026-09-19T23:59:00-12:00' WHERE immichID=?", selectionID(i))
		selectionSQL(t, db, "INSERT INTO albumAssets VALUES (?,'album',?)", testUserID, selectionID(i))
		selectionSQL(t, db, "INSERT INTO assetTags VALUES (?,'tag',?)", testUserID, selectionID(i))
	}
	selectionSQL(t, db, "UPDATE assets SET originalPath='/Trip/Day1/b.jpg',dateTimeOriginal='2026-09-19T00:01:00+14:00',latitude=0 WHERE immichID=?", selectionID(2))
	selectionSQL(t, db, "UPDATE assets SET originalPath='/Trips/c.jpg' WHERE immichID=?", selectionID(3))
	selectionSQL(t, db, "UPDATE assets SET originalPath='/trip/d.jpg' WHERE immichID=?", selectionID(4))
	selectionSQL(t, db, "UPDATE assets SET dateTimeOriginal=NULL WHERE immichID=?", selectionID(5))
	selectionSQL(t, db, "UPDATE assets SET latitude=0,longitude=0 WHERE immichID=?", selectionID(6))
	selectionSQL(t, db, "UPDATE assets SET dateTimeOriginal='2026-09-18T23:59:00' WHERE immichID=?", selectionID(7))
	selectionSQL(t, db, "UPDATE assets SET dateTimeOriginal='2026-02-30T12:00:00' WHERE immichID=?", selectionID(8))
	selectionSQL(t, db, "UPDATE assets SET stackPrimaryAssetID=? WHERE immichID=?", selectionID(1), selectionID(9))
	selectionSQL(t, db, "UPDATE assets SET libraryID='hidden-library' WHERE immichID=?", selectionID(10))
	selectionSQL(t, db, "UPDATE assets SET userID='other' WHERE immichID=?", selectionID(11))
	selectionSQL(t, db, "DELETE FROM assetTags WHERE assetID=?", selectionID(12))
	// An asset has several relations; the requested relation never multiplies it.
	selectionSQL(t, db, "INSERT INTO albumAssets VALUES (?,'second-album',?)", testUserID, selectionID(1))
	selectionSQL(t, db, "INSERT INTO assetTags VALUES (?,'second-tag',?)", testUserID, selectionID(1))
	for _, tc := range []struct {
		name  string
		scope selection.Scope
		ids   []int
	}{
		{"all", selection.Scope{View: "all"}, []int{12, 8, 7, 5, 4, 3, 2, 1}},
		{"album-tag", selection.Scope{View: "album", AlbumID: "album", TagID: "tag", StartDate: "2026-09-19", EndDate: "2026-09-19"}, []int{4, 3, 2, 1}},
		{"folder-tag-date", selection.Scope{View: "folder", FolderPath: "/Trip/", TagID: "tag", StartDate: "2026-09-19", EndDate: "2026-09-19"}, []int{2, 1}},
		{"folder-no-date", selection.Scope{View: "folder", FolderPath: "/Trip", TagID: "tag"}, []int{8, 7, 5, 2, 1}},
		{"gps", selection.Scope{View: "album", AlbumID: "album", TagID: "tag", GPSFilter: "with-gps", StartDate: "2026-09-19"}, []int{6}},
		{"all-gps-end-only", selection.Scope{View: "all", GPSFilter: "all", EndDate: "2026-09-19"}, []int{12, 7, 6, 4, 3, 2, 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := matchingPreview(t, db, tc.scope)
			matchingIDs(t, result, tc.ids...)
			if result.MatchedCount != len(tc.ids) || result.ExcludedCount != 0 {
				t.Fatalf("suppressed candidates leaked into counts: %+v", result)
			}
			loaded, err := selectionStore(t, db).load(context.Background(), testUserID, *result.SnapshotID)
			if err != nil {
				t.Fatal("query facts differ from retained resolver", err)
			}
			matchingIDs(t, loaded, tc.ids...)
		})
	}
}

func TestAISelectionMatchingRejectsAmbiguousHTTPInput(t *testing.T) {
	db, _ := aiTestHandler(t, true)
	h := newAISelectionHandler(db, &Config{AIEnabled: true, ImmichURL: "https://immich.example/api", AIPublicOrigin: aiTestOrigin})
	for _, body := range []string{
		`{"mode":"all-matching","assetIDs":null,"scope":{"view":"all"}}`,
		`{"mode":"all-matching","assetIDs":[],"scope":{"view":"all"}}`,
		`{"mode":"all-matching","page":2,"scope":{"view":"all"}}`,
		`{"mode":"all-matching","cursor":"next","scope":{"view":"all"}}`,
		`{"mode":"all-matching","count":1,"scope":{"view":"all"}}`,
		`{"mode":"all-matching","sort":"name","scope":{"view":"all"}}`,
		`{"mode":"all-matching","scope":{"view":"all","city":"Porto"}}`,
		`{"mode":"all-matching","scope":{"view":"album","albumID":"missing"}}`,
		`{"mode":"all-matching","scope":{"view":"all","tagID":"missing"}}`,
		`{"mode":"all-matching","scope":{"view":"folder","folderPath":"/"}}`,
		`{"mode":"all-matching","scope":{"view":"all","startDate":"2026-09-20","endDate":"2026-09-19"}}`,
	} {
		rec := aiRequest(h, "POST", "/ai/selection-preview", body, aiTestOrigin, true)
		if rec.Code != 400 {
			t.Fatalf("ambiguous input: %d %s", rec.Code, rec.Body.String())
		}
	}
	var count int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&count); err != nil || count != 0 {
		t.Fatalf("invalid request persisted: %d %v", count, err)
	}
}
