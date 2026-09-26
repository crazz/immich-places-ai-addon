package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func approveStackFixture(t *testing.T, f *aiStackFixture, ids []string) writeback.Operation {
	t.Helper()
	ctx := context.Background()
	review, err := f.writer.drafts.observeStack(ctx, testUserID, f.draft.ID, f.draft.Revision, ids, f.image.service)
	if err != nil {
		t.Fatal(err)
	}
	preview, err := (&aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service}).createStack(ctx, testUserID, f.draft.ID, f.draft.Revision, review.ID)
	if err != nil {
		t.Fatal(err)
	}
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "stack-execution"})
	if err != nil {
		t.Fatal(err)
	}
	return op
}

func TestAIStackExecutesExactIndependentTargetsAndRefreshesEach(t *testing.T) {
	f := stackWriteFixture(t)
	addStackFixtureMember(t, f, selectionID(4))
	op := approveStackFixture(t, f, f.members[:3])
	var mu sync.Mutex
	sends := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "private-immich-image-key" {
			t.Error("wrong authority")
			w.WriteHeader(403)
			return
		}
		if r.Method == "GET" {
			if !f.image.handle(w, r) {
				http.NotFound(w, r)
			}
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/assets/")
		var targetIndex = -1
		for i, target := range op.Plan.Manifest.Targets {
			if target.AssetID == id {
				targetIndex = i
			}
		}
		if r.Method != "PATCH" || targetIndex < 0 || r.URL.Path != "/api/assets/"+id {
			t.Error("unapproved mutation")
			w.WriteHeader(400)
			return
		}
		target := op.Plan.Manifest.Targets[targetIndex]
		var payload map[string]any
		if json.NewDecoder(r.Body).Decode(&payload) != nil || payload["latitude"] != target.Intended.Latitude || payload["longitude"] != target.Intended.Longitude {
			t.Error("wrong approved GPS")
		}
		count := 2
		if target.Description != nil {
			count++
			if payload["description"] != target.Description.Intended {
				t.Error("wrong approved description")
			}
		}
		if len(payload) != count {
			t.Error("unapproved fields")
		}
		mu.Lock()
		sends[id]++
		mu.Unlock()
		f.mu.Lock()
		exif := f.metadata[id]["exifInfo"].(map[string]any)
		for key, value := range payload {
			exif[key] = value
		}
		f.mu.Unlock()
		w.WriteHeader(200)
	}))
	t.Cleanup(server.Close)
	f.image.service.endpoint = server.URL
	ctx := context.Background()
	for range 4 {
		if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
			t.Fatal("stack step failed", err)
		}
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || saved.Status != "succeeded" || !saved.Verified || !saved.Refreshed || len(saved.Targets) != 3 {
		t.Fatal("incorrect aggregate", saved.Status, err)
	}
	for _, target := range saved.Targets {
		if target.Status != "succeeded" || target.Attempts != 1 || !target.Verified || !target.Refreshed {
			t.Fatal("incorrect target outcome", target.AssetID)
		}
		for _, field := range target.Fields {
			if field.Status != "verified" || !field.WasVerified {
				t.Fatal("missing verified field")
			}
		}
		var lat, lon sql.NullFloat64
		if err := f.f.db.db.QueryRow(`SELECT latitude,longitude FROM assets WHERE userID=? AND immichID=?`, testUserID, target.AssetID).Scan(&lat, &lon); err != nil || !lat.Valid || !lon.Valid || lat.Float64 != 0 || lon.Float64 != 12 {
			t.Fatal("target refresh failed", err)
		}
		mu.Lock()
		count := sends[target.AssetID]
		mu.Unlock()
		if count != 1 {
			t.Fatal("hidden resend or missing target", count)
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, id := range []string{selectionB, selectionID(3)} {
		if f.metadata[id]["exifInfo"].(map[string]any)["description"] != "sibling private text" {
			t.Fatal("sibling description changed")
		}
	}
	if f.metadata[selectionID(4)]["exifInfo"].(map[string]any)["latitude"] != -12 {
		t.Fatal("unselected member changed")
	}
}
