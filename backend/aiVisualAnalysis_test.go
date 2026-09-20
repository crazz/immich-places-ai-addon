package main

import (
	"context"
	"testing"
)

func TestAIVisualAnalysisUsesExactAuthorizedImageAndProvider(t *testing.T) {
	f := newAIVisualFixture(t)
	result, err := f.analyzer.analyze(context.Background(), f.request)
	if err != nil || result == nil {
		t.Fatal("analysis failed", err)
	}
	doc, err := result.Proposal().Data()
	if err != nil || doc.Outcome != "unknown" {
		t.Fatal("unknown not retained", err)
	}
	if f.hits.Load() != 1 || f.reservations.Load() != 1 || result.Info().Model != "bound-model" {
		t.Fatal("wrong revision/attempt")
	}
	if _, bad := f.image.counts(); bad != 0 {
		t.Fatal("unexpected Immich operation")
	}
}
