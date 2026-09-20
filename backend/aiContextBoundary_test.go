package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"

	"immich-places-backend/internal/ai/contextual"
)

func TestAIContextRejectsObservedChangesAtEveryAttemptBoundary(t *testing.T) {
	for _, stage := range []string{"before", "resolved", "responded"} {
		for _, change := range []string{"capture", "consent", "profile", "installation", "hidden", "source"} {
			t.Run(stage+"/"+change, func(t *testing.T) {
				f := newAIVisualFixture(t)
				var changed atomic.Bool
				f.image.handle = func(w http.ResponseWriter, r *http.Request) bool {
					meta := f.image.metadata()
					capture := "2026-09-20T06:30:00Z"
					if changed.Load() {
						if change == "capture" {
							capture = "2026-09-21T06:30:00Z"
						}
						if change == "source" {
							meta["updatedAt"] = "2026-09-21T12:00:00Z"
						}
					}
					meta["exifInfo"] = map[string]any{"dateTimeOriginal": capture}
					_ = json.NewEncoder(w).Encode(meta)
					return true
				}
				s, req := preparedContextAttempt(t, f)
				req.Consent.Classes = []contextual.Class{contextual.Capture}
				req.Authorize = func(context.Context) error {
					if change == "consent" && changed.Load() {
						return errors.New("revoked private consent")
					}
					return nil
				}
				var err error
				f.request.Context, err = s.prepare(context.Background(), req)
				if err != nil {
					t.Fatal(err)
				}
				mutate := func(ctx context.Context) {
					changed.Store(true)
					var query string
					switch change {
					case "profile":
						query = "UPDATE ai_provider_profiles SET enabled=0"
					case "installation":
						query = "UPDATE ai_installation_identity SET id='changed'"
					case "hidden":
						query = "UPDATE assets SET isHidden=1"
					}
					if query != "" {
						if _, err := f.image.db.db.ExecContext(ctx, query); err != nil {
							t.Error(err)
						}
					}
				}
				switch stage {
				case "before":
					mutate(context.Background())
				case "resolved":
					f.analyzer.dispatcher.AfterResolve = mutate
				case "responded":
					f.handle = func(w http.ResponseWriter, r *http.Request) bool { mutate(r.Context()); return false }
				}
				result, err := f.analyzer.analyzeContext(context.Background(), f.request, s, req)
				if result != nil || err == nil {
					t.Fatal("changed context published")
				}
				want := int32(0)
				if stage == "responded" {
					want = 1
				}
				if f.hits.Load() != want || f.reservations.Load() != want {
					t.Fatal("unexpected transmission", f.hits.Load(), f.reservations.Load())
				}
			})
		}
	}
}
