package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestAIVisualRejectsAuthorityAndSourceChangesAtBothBoundaries(t *testing.T) {
	for _, stage := range []string{"resolved", "responded"} {
		for _, change := range []string{"profile", "policy", "hidden", "source", "credential", "installation", "caller"} {
			t.Run(stage+"/"+change, func(t *testing.T) {
				f := newAIVisualFixture(t)
				var changed atomic.Bool
				f.request.Guard.Authorize = func(context.Context) error {
					if change == "caller" && changed.Load() {
						return errors.New("revoked")
					}
					return nil
				}
				f.image.handle = func(w http.ResponseWriter, r *http.Request) bool {
					if change != "source" || !changed.Load() {
						return false
					}
					meta := f.image.metadata()
					meta["updatedAt"] = "2026-09-21T12:00:00Z"
					_ = json.NewEncoder(w).Encode(meta)
					return true
				}
				mutate := func(ctx context.Context) {
					changed.Store(true)
					var query string
					switch change {
					case "profile":
						query = `UPDATE ai_provider_profiles SET enabled=0`
					case "policy":
						query = `UPDATE ai_provider_capability_checks SET policyFingerprint='changed'`
					case "hidden":
						query = `UPDATE assets SET isHidden=1`
					case "installation":
						query = `UPDATE ai_installation_identity SET id='changed'`
					case "credential":
						query = `UPDATE ai_provider_versions SET secretCiphertext=NULL`
					}
					if query != "" {
						if _, err := f.image.db.db.ExecContext(ctx, query); err != nil {
							t.Error(err)
						}
					}
				}
				if stage == "resolved" {
					f.analyzer.dispatcher.AfterResolve = mutate
				} else {
					f.handle = func(w http.ResponseWriter, r *http.Request) bool { mutate(r.Context()); return false }
				}
				result, err := f.analyzer.analyze(context.Background(), f.request)
				if err == nil || result != nil {
					t.Fatal("stale authority published")
				}
				want := int32(0)
				if stage == "responded" {
					want = 1
				}
				if f.hits.Load() != want || f.reservations.Load() != want {
					t.Fatal("unexpected dispatch", f.hits.Load(), f.reservations.Load())
				}
			})
		}
	}
}
