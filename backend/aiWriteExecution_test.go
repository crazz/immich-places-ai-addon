package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIWriteExecutesReservedExactGPSDespiteChangedStack(t *testing.T) {
	f, image, store, draft, meta := writePreviewFixture(t)
	p := savedWritePreview(t, f, image, draft)
	var mu sync.Mutex
	sends := 0
	meta["stack"] = map[string]any{"id": selectionID(22), "primaryAssetId": selectionB, "assetCount": 6}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.URL.Path != "/api/assets/"+draft.AssetID {
			t.Error("expanded target", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Method == "PATCH" {
			var reserved int
			if err := f.db.db.QueryRow("SELECT attempts FROM ai_write_targets").Scan(&reserved); err != nil || reserved != 1 {
				t.Error("not reserved", reserved, err)
			}
			var body map[string]float64
			if json.NewDecoder(r.Body).Decode(&body) != nil || len(body) != 2 || body["latitude"] != 0 || body["longitude"] != 12 {
				t.Error("unexpected fields", body)
			}
			sends++
			meta["exifInfo"] = body
		}
		if err := json.NewEncoder(w).Encode(meta); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	image.service.endpoint = server.URL
	writer := &aiWriteStore{drafts: store, enabled: func() bool { return true }, profile: "immich-v3.2.2", images: image.service, sync: newSyncService(f.db, nil, nil)}
	op, err := writer.confirm(context.Background(), testUserID, writeback.Confirmation{PreviewID: p.Plan.ID, Digest: p.Digest, Key: "execute"})
	if err != nil {
		t.Fatal(err)
	}
	if err = writer.runOne(context.Background(), testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	op, err = writer.get(context.Background(), testUserID, op.ID, false)
	if err != nil || op.Status != "succeeded" || !op.Verified || !op.Refreshed || op.Attempts != 1 || op.Noop {
		t.Fatal(op, err)
	}
	mu.Lock()
	count := sends
	mu.Unlock()
	if count != 1 {
		t.Fatal("sends", count)
	}
	row, err := f.db.getAssetByID(context.Background(), testUserID, draft.AssetID)
	if err != nil || row.Latitude == nil || row.Longitude == nil || *row.Latitude != 0 || *row.Longitude != 12 {
		t.Fatal(row, err)
	}
}
