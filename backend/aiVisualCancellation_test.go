package main

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestAIVisualCancellationDoesNotPublishOrLeakAuthorityFailure(t *testing.T) {
	for _, stage := range []string{"before", "metadata", "provider"} {
		t.Run(stage, func(t *testing.T) {
			f := newAIVisualFixture(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			switch stage {
			case "before":
				cancel()
			case "metadata":
				f.image.handle = func(w http.ResponseWriter, r *http.Request) bool {
					cancel()
					http.Error(w, "canceled", 503)
					return true
				}
			case "provider":
				f.handle = func(w http.ResponseWriter, r *http.Request) bool { cancel(); return false }
			}
			result, err := f.analyzer.analyze(ctx, f.request)
			if result != nil || !errors.Is(err, context.Canceled) {
				t.Fatal("cancellation misclassified", err)
			}
			want := int32(0)
			if stage == "provider" {
				want = 1
			}
			if f.hits.Load() != want {
				t.Fatal("unexpected dispatch")
			}
		})
	}
}
