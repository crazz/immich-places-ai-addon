package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIJobWorkerPersistsSafeFailureWithoutHiddenRetry(t *testing.T) {
	for _, kind := range []string{"transient", "permanent", "blocked", "raw", "invalid", "budget"} {
		t.Run(kind, func(t *testing.T) {
			f := newAIJobFixture(t)
			ctx := context.Background()
			f.input.AssetIDs = []string{selectionA}
			f.input.MaxCalls = 1
			if kind != "budget" {
				f.input.MaxCalls = 3
			}
			job, err := f.store.Submit(ctx, f.input)
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			worker := jobs.Worker{Store: f.store, Policy: jobs.DefaultPolicy(), Jitter: func() time.Duration { return time.Second }, Execute: func(ctx context.Context, lease jobs.Lease, guard jobs.Guard) (jobs.Completion, error) {
				if err := guard.Reserve(ctx); err != nil {
					return jobs.Completion{}, err
				}
				calls++
				switch kind {
				case "transient":
					return jobs.Completion{}, &jobs.Failure{Code: jobs.Transient}
				case "permanent":
					return jobs.Completion{}, jobs.Failure{Code: jobs.Permanent}
				case "blocked":
					return jobs.Completion{}, &jobs.Failure{Code: jobs.Blocked}
				case "raw":
					return jobs.Completion{}, errors.New("private-provider-key")
				case "budget":
					return jobs.Completion{}, guard.Reserve(ctx)
				}
				return jobs.Completion{}, nil
			}}
			ran, err := worker.RunOne(ctx, make(chan time.Time))
			if err != nil || !ran || calls != 1 {
				t.Fatal("worker failure not finalized", ran, calls, err)
			}
			got, err := f.store.Get(ctx, testUserID, job.ID)
			if err != nil {
				t.Fatal(err)
			}
			want := "failed"
			if kind == "transient" {
				want = "retry_wait"
			}
			if kind == "blocked" {
				want = "blocked"
			}
			if got.Items[0].State != want || got.Calls != 1 || strings.Contains(got.Items[0].Failure, "private") {
				t.Fatal("unsafe failure state", got.Items, got.Calls)
			}
			if kind == "budget" && got.Items[0].Failure != "budget" {
				t.Fatal("budget category lost")
			}
		})
	}
}
