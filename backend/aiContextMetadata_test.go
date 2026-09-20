package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"immich-places-backend/internal/ai/contextual"
)

func TestAIContextRejectsAmbiguousCaptureMetadata(t *testing.T) {
	f := newAIImageFixture(t)
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		base, _ := json.Marshal(f.metadata())
		_, _ = fmt.Fprintf(w, `%s,"exifInfo":{"dateTimeOriginal":"2026-09-20T06:30:00Z","DateTimeOriginal":"2026-09-21T06:30:00Z"}}`, base[:len(base)-1])
		return true
	}
	req := contextRequest(t, f)
	req.Consent.Classes = []contextual.Class{contextual.Capture}
	if b, err := (&aiContextPreparer{images: f.service}).prepare(context.Background(), req); err == nil || b != nil {
		t.Fatal("ambiguous capture accepted")
	}
}
