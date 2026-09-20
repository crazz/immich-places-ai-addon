package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"immich-places-backend/internal/ai/contextual"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/selection"
)

func TestAIResearchFreezesDisplayedAlbumAndCaptureBeforeCatalogChanges(t *testing.T) {
	f, p, req := researchProductionFixture(t)
	ctx := context.Background()
	album := selectionID(80)
	selectionSQL(t, f.image.db, "UPDATE assets SET dateTimeOriginal='2009-04-01T12:30:00' WHERE immichID=?", selectionA)
	selectionSQL(t, f.image.db, "INSERT INTO albums(userID,immichID,albumName,updatedAt) VALUES(?,?,?,'')", testUserID, album, "Israel trip")
	selectionSQL(t, f.image.db, "INSERT INTO albumAssets(userID,albumID,assetID) VALUES(?,?,?)", testUserID, album, selectionA)
	preview, err := p.selections.preview(ctx, selection.Input{Mode: "explicit", AssetIDs: []string{selectionA}, Scope: &selection.Scope{View: "album", AlbumID: album}}, testUserID)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(preview)
	if !strings.Contains(string(raw), "Israel trip") || !strings.Contains(string(raw), "2009-04-01T12:30:00") {
		t.Fatal("optional values not shown in preview")
	}
	req.Configuration.SelectionToken = *preview.SnapshotID
	req.Configuration.Context.Classes = []contextual.Class{contextual.Hint, contextual.AlbumLabel, contextual.Capture}
	req.Configuration.Context.AlbumID = album
	req.Consent.Configuration = req.Configuration
	selectionSQL(t, f.image.db, "UPDATE albums SET albumName='Changed after preview'")
	selectionSQL(t, f.image.db, "UPDATE assets SET dateTimeOriginal='2015-01-01T00:00:00'")
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	original := f.handle
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		messages := string(body["messages"])
		if !strings.Contains(messages, "Israel trip") || !strings.Contains(messages, "2009-04-01") || strings.Contains(messages, "Changed after preview") || strings.Contains(messages, "2015-01-01") {
			t.Error("displayed inputs changed before dispatch")
		}
		encoded, _ := json.Marshal(body)
		r.Body = io.NopCloser(strings.NewReader(string(encoded)))
		return original(w, r)
	}
	worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
	if _, err := worker.RunOne(ctx, make(chan time.Time)); err != nil {
		t.Fatal(err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Counts["succeeded"] != 1 {
		t.Fatal("frozen Research inputs failed", progress, err)
	}
}
