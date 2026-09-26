package main

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStackRetryRejectsUnknownStaleExpiredExhaustedAndCoalescesConcurrent(t *testing.T) {
	for _, scenario := range []string{"unknown", "stale", "expired", "exhausted", "concurrent"} {
		t.Run(scenario, func(t *testing.T) {
			f := stackWriteFixture(t)
			op := approveStackFixture(t, f, f.members)
			asset := op.Targets[1].AssetID
			m := stackMutationServer(t, f)
			m.handle = func(w http.ResponseWriter, r *http.Request) bool {
				if r.Method == "PATCH" && r.URL.Path == "/api/assets/"+asset {
					w.WriteHeader(400)
					return true
				}
				return false
			}
			ctx := context.Background()
			if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
				t.Fatal(err)
			}
			if scenario == "unknown" {
				step, err := writeback.TargetStep(op, asset)
				if err != nil {
					t.Fatal(err)
				}
				a := &aiStackAttempt{store: f.writer, parent: op, assetID: asset}
				if _, err := a.Read(ctx, step); err != nil {
					t.Fatal(err)
				}
				if err := a.Reserve(ctx, step, false); err != nil {
					t.Fatal(err)
				}
				_ = a.Send(ctx, step)
				if err := a.Sent(ctx, step, writeback.Completion{Known: false, Code: "response_lost"}); err != nil {
					t.Fatal(err)
				}
				if err := a.Verify(ctx, step); err != nil {
					t.Fatal(err)
				}
			} else if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
				t.Fatal(err)
			}
			if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
				t.Fatal(err)
			}
			saved, err := f.writer.get(ctx, testUserID, op.ID, false)
			if err != nil || saved.Targets[0].Status != "succeeded" || saved.Targets[2].Status != "succeeded" {
				t.Fatal("fixture siblings not complete", err)
			}
			generation := saved.Targets[1].Generation
			wantAttempts := 1
			switch scenario {
			case "stale":
				generation++
			case "expired":
				f.f.now = f.f.now.Add(6 * time.Minute)
			case "exhausted", "concurrent":
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
						accepted, err := f.writer.retryStackTarget(ctx, testUserID, op.ID, asset, generation)
						if err != nil || accepted.Targets[1].Generation != generation+1 || accepted.Targets[1].Attempts != 1 {
							t.Error("same retry generation did not coalesce", err)
						}
					}()
				}
				close(start)
				group.Wait()
				if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
					t.Fatal(err)
				}
				saved, err = f.writer.get(ctx, testUserID, op.ID, false)
				if err != nil || saved.Targets[1].Status != "failed" || saved.Targets[1].Code != "ATTEMPTS_EXHAUSTED" {
					t.Fatal("second unsuccessful attempt did not exhaust budget", err)
				}
				generation, wantAttempts = saved.Targets[1].Generation, 2
			}
			if _, err := f.writer.retryStackTarget(ctx, testUserID, op.ID, asset, generation); err == nil {
				t.Fatal("ineligible target gained another retry")
			}
			f.f.reopen(t)
			f.writer.drafts.results.jobs = f.f.store
			saved, err = f.writer.get(ctx, testUserID, op.ID, false)
			if err != nil || saved.Targets[1].Attempts != wantAttempts || m.sent(asset) != wantAttempts {
				t.Fatal("retry budget or sends changed on rejection/restart", err)
			}
			for _, index := range []int{0, 2} {
				if saved.Targets[index].Status != "succeeded" || saved.Targets[index].Attempts != 1 || m.sent(saved.Targets[index].AssetID) != 1 {
					t.Fatal("retry action changed successful sibling")
				}
			}
		})
	}
}
