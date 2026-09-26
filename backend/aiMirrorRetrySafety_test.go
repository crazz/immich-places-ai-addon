package main

import (
	"context"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorRetryRejectsStaleExpiredUnavailableExhaustedAndCoalescesConcurrent(t *testing.T) {
	for _, scenario := range []string{"stale", "expired", "unavailable", "exhausted", "concurrent", "accepted-replay"} {
		t.Run(scenario, func(t *testing.T) {
			f := mirrorWriteFixture(t)
			mirrorMutationServer(t, f)
			ctx := context.Background()
			f.handle = func(w http.ResponseWriter, r *http.Request) bool {
				if r.Method == "PUT" {
					f.mu.Lock()
					f.mirrorSends++
					f.mu.Unlock()
					w.WriteHeader(400)
					return true
				}
				return false
			}
			op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "retry-safety"})
			if err != nil {
				t.Fatal(err)
			}
			for range 2 {
				if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
					t.Fatal(err)
				}
			}
			saved, err := f.writer.get(ctx, testUserID, op.ID, false)
			if err != nil || saved.Mirror.Status != "retryable" {
				t.Fatal("fixture did not produce known unsuccessful metadata", err)
			}
			generation := saved.Mirror.Generation
			acceptedGeneration := generation
			wantAttempts := 1
			switch scenario {
			case "stale":
				generation++
			case "expired":
				f.f.now = f.f.now.Add(6 * time.Minute)
			case "unavailable":
				f.handle = func(w http.ResponseWriter, r *http.Request) bool { w.WriteHeader(503); return true }
			default:
				requests := 1
				if scenario == "concurrent" {
					requests = 2
				}
				start := make(chan struct{})
				var group sync.WaitGroup
				for range requests {
					group.Add(1)
					go func() {
						defer group.Done()
						<-start
						accepted, err := f.writer.retryMirror(ctx, testUserID, op.ID, f.draft.AssetID, generation)
						if err != nil || accepted.Mirror.Generation != generation+1 || accepted.Mirror.Attempts != 1 {
							t.Error("same generation did not coalesce", err)
						}
					}()
				}
				close(start)
				group.Wait()
				if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
					t.Fatal(err)
				}
				saved, err = f.writer.get(ctx, testUserID, op.ID, false)
				if err != nil || saved.Mirror.Status != "failed" || saved.Mirror.Attempts != 2 {
					t.Fatal("second unsuccessful metadata attempt did not exhaust budget", err)
				}
				generation = saved.Mirror.Generation
				wantAttempts = 2
			}
			if _, err = f.writer.retryMirror(ctx, testUserID, op.ID, f.draft.AssetID, generation); err == nil {
				t.Fatal("ineligible metadata retry accepted")
			}
			f.f.reopen(t)
			f.writer.drafts.results.jobs = f.f.store
			retained, err := f.writer.get(ctx, testUserID, op.ID, false)
			if err != nil || retained.Mirror.Attempts != wantAttempts || f.mirrorSends != wantAttempts || f.standardSends != 1 {
				t.Fatal("rejection/reopen changed attempts or standard step", err)
			}
			if scenario == "accepted-replay" {
				f.f.now = f.f.now.Add(6 * time.Minute)
				f.writer.enabled = func() bool { return false }
				f.handle = func(w http.ResponseWriter, r *http.Request) bool {
					t.Error("accepted replay performed upstream I/O")
					w.WriteHeader(503)
					return true
				}
				replay, err := f.writer.retryMirror(ctx, testUserID, op.ID, f.draft.AssetID, acceptedGeneration)
				if err != nil || !reflect.DeepEqual(retained, replay) {
					t.Fatal("expired disabled accepted replay lost durable identity", err)
				}
				if _, err = f.writer.retryMirror(ctx, "foreign-owner", op.ID, f.draft.AssetID, acceptedGeneration); err == nil {
					t.Fatal("accepted replay disclosed foreign history")
				}
				selectionSQL(t, f.f.db, `UPDATE ai_installation_identity SET id='replacement' WHERE singleton=1`)
				if _, err = f.writer.retryMirror(ctx, testUserID, op.ID, f.draft.AssetID, acceptedGeneration); err == nil {
					t.Fatal("accepted replay crossed installation binding")
				}
			}
		})
	}
}
