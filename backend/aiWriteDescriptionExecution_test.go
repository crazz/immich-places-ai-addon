package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIDescriptionOnlyDispatchWritesOneExactFieldAndPreservesGPS(t *testing.T) {
	w := standardWriteFixture(t, []string{"description"})
	ctx := context.Background()
	op, err := w.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "exact-description"})
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	sends := 0
	w.meta["exifInfo"] = map[string]any{"latitude": 35, "longitude": 22, "description": w.preview.Plan.Description.Before.Value}
	server := httptest.NewServer(http.HandlerFunc(func(out http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.URL.Path != "/api/assets/"+w.draft.AssetID || r.Header.Get("x-api-key") != "private-immich-image-key" {
			out.WriteHeader(403)
			return
		}
		if r.Method == "PATCH" {
			sends++
			var payload map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || len(payload) != 1 {
				t.Error("unselected fields sent", err)
				out.WriteHeader(400)
				return
			}
			var text string
			if json.Unmarshal(payload["description"], &text) != nil || text != w.preview.Plan.Description.Intended {
				t.Error("exact reviewed description changed")
				out.WriteHeader(400)
				return
			}
			w.meta["exifInfo"] = map[string]any{"latitude": 35, "longitude": 22, "description": text}
		} else if r.Method != "GET" {
			t.Error("unexpected method")
			out.WriteHeader(400)
			return
		}
		if err := json.NewEncoder(out).Encode(w.meta); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	w.image.service.endpoint = server.URL
	if err := w.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	finished, err := w.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || finished.Status != "succeeded" || !finished.Verified || finished.Refreshed || !finished.Settled {
		t.Fatalf("description outcome: %+v %v", finished, err)
	}
	mu.Lock()
	count := sends
	mu.Unlock()
	if count != 1 {
		t.Fatal("wrong mutation count", count)
	}
	var latitude, longitude sql.NullFloat64
	if err := w.f.db.db.QueryRow(`SELECT latitude,longitude FROM assets WHERE userID=? AND immichID=?`, testUserID, w.draft.AssetID).Scan(&latitude, &longitude); err != nil || latitude.Valid || longitude.Valid {
		t.Fatal("description-only write changed Missing GPS membership", err)
	}
	raw, err := json.Marshal(finished)
	var projection struct {
		Fields []struct{ Field, Status string } `json:"fields"`
	}
	if err != nil || json.Unmarshal(raw, &projection) != nil || len(projection.Fields) != 1 || projection.Fields[0].Field != "description" || projection.Fields[0].Status != "verified" {
		t.Fatal("durable field outcome missing")
	}
}
