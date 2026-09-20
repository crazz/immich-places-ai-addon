package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/contextual"
)

func TestAIContextUsesOnlyCurrentEligibleMetadataNeighbors(t *testing.T) {
	f := newAIImageFixture(t)
	req := contextRequest(t, f)
	req.Consent.Classes = []contextual.Class{contextual.Capture, contextual.Neighbors}
	for _, id := range []string{selectionB, selectionID(3), selectionID(4)} {
		seedAsset(t, f.db, id, ptr(1.0), ptr(2.0), "2026-09-20T06:31:00Z")
		if _, err := f.db.db.Exec("UPDATE assets SET dateTimeOriginal=? WHERE immichID=?", "2026-09-20T06:31:00Z", id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.db.db.Exec("UPDATE assets SET isHidden=1 WHERE immichID=?", selectionID(4)); err != nil {
		t.Fatal(err)
	}
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		id := strings.TrimPrefix(r.URL.Path, "/api/assets/")
		meta := f.metadata()
		meta["id"] = id
		meta["exifInfo"] = map[string]any{"dateTimeOriginal": "2026-09-20T06:31:00Z", "latitude": 1, "longitude": 2}
		if id == selectionA {
			meta["exifInfo"] = map[string]any{"dateTimeOriginal": "2026-09-20T06:30:00Z"}
		}
		_ = json.NewEncoder(w).Encode(meta)
		return true
	}
	s := &aiContextPreparer{images: f.service, lineage: func(_ context.Context, owner, id string) (string, error) {
		if owner != testUserID {
			t.Fatal("foreign lineage lookup")
		}
		if id == selectionID(3) {
			return "ai", nil
		}
		return "unknown", nil
	}}
	b, err := s.prepare(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	p, _ := b.Projection()
	var projection contextual.Projection
	_ = json.Unmarshal(p, &projection)
	if len(projection.Sources) != 2 || projection.Sources[1].Kind != contextual.Neighbors {
		t.Fatalf("eligible neighbor missing: %s", p)
	}
	if b.Info().Sources[1].Asset != selectionB || b.Info().Sources[1].Lineage != "unknown" {
		t.Fatal("neighbor provenance lost")
	}
	for _, path := range f.paths {
		if strings.Contains(path, "thumbnail") || strings.Contains(path, selectionID(3)) || strings.Contains(path, selectionID(4)) {
			t.Fatal("ineligible/image read", path)
		}
	}
}

func TestAIContextOmitsUnavailableOptionalNeighborBeforeFreezing(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusNotFound} {
		f := newAIImageFixture(t)
		req := contextRequest(t, f)
		req.Consent.Classes = []contextual.Class{contextual.Neighbors}
		seedAsset(t, f.db, selectionB, ptr(1.0), ptr(2.0), "2026-09-20T06:31:00Z")
		selectionSQL(t, f.db, "UPDATE assets SET dateTimeOriginal=? WHERE immichID=?", "2026-09-20T06:31:00Z", selectionB)
		f.handle = func(w http.ResponseWriter, r *http.Request) bool {
			if r.URL.Path == "/api/assets/"+selectionB {
				http.Error(w, "private unavailable source", status)
				return true
			}
			meta := f.metadata()
			meta["exifInfo"] = map[string]any{"dateTimeOriginal": "2026-09-20T06:30:00Z"}
			_ = json.NewEncoder(w).Encode(meta)
			return true
		}
		b, err := (&aiContextPreparer{images: f.service}).prepare(context.Background(), req)
		if err != nil || b == nil || len(b.Info().Sources) != 0 || len(b.Info().Omissions) == 0 {
			t.Fatal("optional unavailable source did not produce empty context", status, err)
		}
	}
}
