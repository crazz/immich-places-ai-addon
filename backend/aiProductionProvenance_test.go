package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/selection"
)

func TestAIProductionRetainsLocalCaptureAndSelectedAlbumProvenance(t *testing.T) {
	f, p, req := productionFixture(t)
	ctx := context.Background()
	album := selectionID(80)
	selectionSQL(t, f.image.db, "UPDATE assets SET dateTimeOriginal='2026-09-19T23:30:00-07:00' WHERE immichID=?", selectionA)
	selectionSQL(t, f.image.db, "INSERT INTO albums(userID,immichID,albumName,updatedAt) VALUES(?,?,?,'')", testUserID, album, "Selected trip")
	selectionSQL(t, f.image.db, "INSERT INTO albumAssets(userID,albumID,assetID) VALUES(?,?,?)", testUserID, album, selectionA)
	snapshot, err := p.selections.preview(ctx, selection.Input{Mode: "explicit", AssetIDs: []string{selectionA}, Scope: &selection.Scope{View: "album", AlbumID: album}}, testUserID)
	if err != nil {
		t.Fatal(err)
	}
	req.Configuration.SelectionToken = *snapshot.SnapshotID
	req.Consent.Configuration = req.Configuration
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.image.db, "UPDATE assets SET dateTimeOriginal=NULL")
	selectionSQL(t, f.image.db, "UPDATE albums SET albumName='Changed later'")
	var capture, albumID, label string
	err = f.image.db.db.QueryRow("SELECT captureTime,albumID,albumLabel FROM ai_job_launch WHERE jobID=? AND assetID=?", job.ID, selectionA).Scan(&capture, &albumID, &label)
	if err != nil || capture != "2026-09-19T23:30:00-07:00" || albumID != album || label != "Selected trip" {
		t.Fatal("launch provenance not retained", capture, albumID, label, err)
	}
	if f.hits.Load() != 0 {
		t.Fatal("local provenance contacted provider")
	}
}
