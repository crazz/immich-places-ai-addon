package main

import (
	"context"
	"net/http"
	"sync"
	"testing"
)

func TestAIVisualConcurrentAttemptsNeverWriteDatabaseOrImmich(t *testing.T) {
	for _, outcome := range []string{"success", "failure", "canceled", "stale-capability"} {
		t.Run(outcome, func(t *testing.T) {
			f := newAIVisualFixture(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if outcome == "failure" {
				f.handle = func(w http.ResponseWriter, r *http.Request) bool {
					http.Error(w, "private-provider-body", 503)
					return true
				}
			}
			if outcome == "canceled" {
				cancel()
			}
			if outcome == "stale-capability" {
				if _, err := f.image.db.db.Exec(`UPDATE ai_provider_capability_checks SET lifecycle='running',deadlineAt='2020-01-01T00:00:00Z'`); err != nil {
					t.Fatal(err)
				}
			}
			f.image.db.db.SetMaxOpenConns(1)
			if _, err := f.image.db.db.Exec(`PRAGMA query_only=ON`); err != nil {
				t.Fatal(err)
			}
			var before int
			if err := f.image.db.db.QueryRow(`SELECT total_changes()`).Scan(&before); err != nil {
				t.Fatal(err)
			}
			var wg sync.WaitGroup
			for range 8 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					result, err := f.analyzer.analyze(ctx, f.request)
					if outcome == "success" {
						if err != nil || result == nil {
							t.Error("successful attempt failed", err)
						}
					} else if err == nil || result != nil {
						t.Error("failed attempt published")
					}
				}()
			}
			wg.Wait()
			var after int
			if err := f.image.db.db.QueryRow(`SELECT total_changes()`).Scan(&after); err != nil {
				t.Fatal(err)
			}
			if before != after {
				t.Fatal("database changed")
			}
			want := int32(0)
			if outcome == "success" || outcome == "failure" {
				want = 8
			}
			if f.hits.Load() != want || f.reservations.Load() != want {
				t.Fatal("unexpected attempt count", f.hits.Load(), f.reservations.Load())
			}
			if _, bad := f.image.counts(); bad != 0 {
				t.Fatal("unexpected Immich operation")
			}
		})
	}
}
