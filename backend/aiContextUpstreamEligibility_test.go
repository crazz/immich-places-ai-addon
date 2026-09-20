package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"immich-places-backend/internal/ai/contextual"
)

func TestAIContextOmitsNeighborThatBecameHiddenUpstream(t *testing.T) {
	f := newAIImageFixture(t)
	seedAsset(t, f.db, selectionB, ptr(1.0), ptr(2.0), "2026-09-20T06:31:00Z")
	selectionSQL(t, f.db, "UPDATE assets SET dateTimeOriginal=? WHERE immichID=?", "2026-09-20T06:31:00Z", selectionB)
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		meta := f.metadata()
		meta["exifInfo"] = map[string]any{"dateTimeOriginal": "2026-09-20T06:30:00Z", "latitude": 1, "longitude": 2}
		if r.URL.Path == "/api/assets/"+selectionB {
			meta["id"], meta["visibility"] = selectionB, "hidden"
		}
		_ = json.NewEncoder(w).Encode(meta)
		return true
	}
	req := contextRequest(t, f)
	req.Consent.Classes = []contextual.Class{contextual.Neighbors}
	b, err := (&aiContextPreparer{images: f.service}).prepare(context.Background(), req)
	if err != nil || b == nil || len(b.Info().Sources) != 0 {
		t.Fatal("hidden optional source not omitted", err)
	}
}
