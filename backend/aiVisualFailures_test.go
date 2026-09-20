package main

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"immich-places-backend/internal/ai/analysis"
)

func TestAIVisualSourceFailureRemainsTransientAndNeverDispatches(t *testing.T) {
	f := newAIVisualFixture(t)
	f.image.handle = func(w http.ResponseWriter, r *http.Request) bool {
		http.Error(w, "private upstream marker", 503)
		return true
	}
	result, err := f.analyzer.analyze(context.Background(), f.request)
	if !errors.Is(err, analysis.ErrUpstream) || result != nil || f.hits.Load() != 0 || f.reservations.Load() != 0 {
		t.Fatal("transient source failure misclassified", err)
	}
}
