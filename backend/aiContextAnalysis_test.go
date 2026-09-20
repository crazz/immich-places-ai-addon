package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/results"
)

func preparedContextAttempt(t *testing.T, f *aiVisualFixture) (*aiContextPreparer, aiContextRequest) {
	t.Helper()
	req := contextRequest(t, f.image)
	req.Binding.Profile, req.Binding.Revision = f.request.ProfileID, f.request.Revision
	req.Consent.Binding = req.Binding
	s := &aiContextPreparer{images: f.image.service}
	var err error
	f.request.Context, err = s.prepare(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	return s, req
}

func TestAIContextAnalyzesThroughCurrentProfileAndSourceAuthority(t *testing.T) {
	f := newAIVisualFixture(t)
	s, req := preparedContextAttempt(t, f)
	var before int
	if err := f.image.db.db.QueryRow("SELECT total_changes()").Scan(&before); err != nil {
		t.Fatal(err)
	}
	result, err := f.analyzer.analyzeContext(context.Background(), f.request, s, req)
	if err != nil || result == nil {
		t.Fatal("context adapter failed", err)
	}
	if result.Info().Mode != results.ContextAssisted || f.hits.Load() != 1 || f.reservations.Load() != 1 {
		t.Fatal("context attempt counts/metadata")
	}
	var after int
	if err := f.image.db.db.QueryRow("SELECT total_changes()").Scan(&after); err != nil || before != after {
		t.Fatal("context mutated storage", err)
	}
}
