package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIJobWorkerCompletesSyntheticAttemptThroughDurableGuards(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	completion := aiJobCompletion(t)
	calls := 0
	worker := jobs.Worker{Store: f.store, Policy: jobs.DefaultPolicy(), Execute: func(ctx context.Context, lease jobs.Lease, guard jobs.Guard) (jobs.Completion, error) {
		if lease.Input.Profile != "profile" || lease.Input.Revision != 1 || lease.Model != "manual-model" {
			t.Error("wrong execution binding")
		}
		if err := guard.Authorize(ctx); err != nil {
			return jobs.Completion{}, err
		}
		if err := guard.Reserve(ctx); err != nil {
			return jobs.Completion{}, err
		}
		calls++
		return completion, nil
	}}
	ran, err := worker.RunOne(ctx, make(chan time.Time))
	if err != nil || !ran || calls != 1 {
		t.Fatal("worker", ran, calls, err)
	}
	got, err := f.store.Get(ctx, testUserID, job.ID)
	if err != nil || got.Items[0].State != "succeeded" || got.Items[1].State != "queued" || got.Calls != 1 {
		t.Fatal("worker durable outcome", got.Items, err)
	}
}
