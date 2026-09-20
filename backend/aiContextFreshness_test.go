package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"immich-places-backend/internal/ai/contextual"
)

func TestAIContextRejectsFrozenCaptureChangesWithoutReplacingBundle(t *testing.T) {
	f := newAIImageFixture(t)
	capture := "2026-09-20T06:30:00Z"
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		meta := f.metadata()
		meta["exifInfo"] = map[string]any{"dateTimeOriginal": capture}
		_ = json.NewEncoder(w).Encode(meta)
		return true
	}
	req := contextRequest(t, f)
	req.Consent.Classes = []contextual.Class{contextual.Capture}
	s := &aiContextPreparer{images: f.service}
	b, err := s.prepare(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	digest := b.Info().Digest
	if err := s.check(context.Background(), req, b); err != nil {
		t.Fatal("unchanged context rejected", err)
	}
	capture = "2026-09-21T06:30:00Z"
	if err := s.check(context.Background(), req, b); err == nil {
		t.Fatal("changed capture accepted")
	}
	if b.Info().Digest != digest {
		t.Fatal("frozen bundle substituted")
	}
}
