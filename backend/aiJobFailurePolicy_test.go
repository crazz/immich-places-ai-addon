package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIJobTransientFailureSchedulesBoundedRetry(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	f.input.AssetIDs = []string{selectionA}
	f.input.MaxCalls = 3
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	lease, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok {
		t.Fatal(err)
	}
	if err = f.store.Reserve(ctx, lease); err != nil {
		t.Fatal(err)
	}
	if err = f.store.Fail(ctx, lease, jobs.Failure{Code: jobs.Transient, RetryAfter: time.Hour}, time.Hour); err != nil {
		t.Fatal(err)
	}
	got, err := f.store.Get(ctx, testUserID, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	due := f.now.Add(5 * time.Minute)
	if got.Items[0].State != "retry_wait" || got.Items[0].Failure != "transient" || got.Items[0].Calls != 1 || got.Items[0].NextAttemptAt != due.UnixNano() {
		t.Fatal("retry policy", got.Items)
	}
	if _, ok, err = f.store.Claim(ctx, jobs.DefaultPolicy()); err != nil || ok {
		t.Fatal("early retry", err)
	}
	f.now = due
	next, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok || next.Attempts != 2 || next.Calls != 1 {
		t.Fatal("due attempt", err)
	}
}

func TestAIJobPermanentFailureDoesNotStopUnrelatedItems(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	first, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok {
		t.Fatal(err)
	}
	if err = f.store.Reserve(ctx, first); err != nil {
		t.Fatal(err)
	}
	if err = f.store.Fail(ctx, first, jobs.Failure{Code: jobs.Permanent}, 0); err != nil {
		t.Fatal(err)
	}
	second, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok || second.ItemID == first.ItemID {
		t.Fatal("unrelated work stopped", err)
	}
	if err = f.store.Reserve(ctx, second); err != nil {
		t.Fatal(err)
	}
	if _, err = f.store.Complete(ctx, second, aiJobCompletion(t)); err != nil {
		t.Fatal(err)
	}
	got, err := f.store.Get(ctx, testUserID, job.ID)
	if err != nil || got.Items[0].State != "failed" || got.Items[1].State != "succeeded" || got.Calls != 2 {
		t.Fatal("failure leaked", got.Items, err)
	}
}

func TestAIJobSystemicFailureBlocksCallsAndLatePublication(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	f.input.AssetIDs = append(f.input.AssetIDs, "33333333-3333-4333-8333-333333333333")
	f.input.MaxCalls = 9
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	policy := jobs.DefaultPolicy()
	policy.PerOwner = 2
	first, ok, err := f.store.Claim(ctx, policy)
	if err != nil || !ok {
		t.Fatal(err)
	}
	second, ok, err := f.store.Claim(ctx, policy)
	if err != nil || !ok {
		t.Fatal(err)
	}
	for _, lease := range []jobs.Lease{first, second} {
		if err = f.store.Reserve(ctx, lease); err != nil {
			t.Fatal(err)
		}
	}
	if err = f.store.Fail(ctx, first, jobs.Failure{Code: jobs.Blocked}, 0); err != nil {
		t.Fatal(err)
	}
	if f.store.Reserve(ctx, second) == nil || f.store.Authorize(ctx, second) == nil {
		t.Fatal("blocked job admitted more work")
	}
	if _, err = f.store.Complete(ctx, second, aiJobCompletion(t)); err == nil {
		t.Fatal("blocked job published late response")
	}
	if err = f.store.Fail(ctx, second, jobs.Failure{Code: jobs.Transient}, 0); err != nil {
		t.Fatal(err)
	}
	got, err := f.store.Get(ctx, testUserID, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range got.Items {
		if item.State != "blocked" {
			t.Fatal("systemic failure retried", got.Items)
		}
	}
	if _, ok, err = f.store.Claim(ctx, policy); err != nil || ok {
		t.Fatal("blocked item claimed", err)
	}
	input := f.input
	input.Key = "unrelated-job"
	if _, err = f.store.Submit(ctx, input); err != nil {
		t.Fatal(err)
	}
	next, ok, err := f.store.Claim(ctx, policy)
	if err != nil || !ok || next.JobID == job.ID {
		t.Fatal("unrelated job blocked", err)
	}
}
