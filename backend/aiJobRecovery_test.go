package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIJobRestartRecoveryFencesOldLeaseAndReservation(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	f.input.AssetIDs = []string{selectionA}
	f.input.MaxCalls = 3
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	old, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok {
		t.Fatal(err)
	}
	if err = f.store.Reserve(ctx, old); err != nil {
		t.Fatal(err)
	}
	f.reopen(t)
	if count, err := f.store.Recover(ctx); err != nil || count != 0 {
		t.Fatal("unexpired work stolen", count, err)
	}
	f.now = old.ExpiresAt
	if count, err := f.store.Recover(ctx); err != nil || count != 1 {
		t.Fatal("expired work not recovered", count, err)
	}
	recovered, err := f.store.Get(ctx, testUserID, job.ID)
	if err != nil || recovered.Items[0].State != "retry_wait" || recovered.Items[0].Calls != 1 {
		t.Fatal("recovery lost state", err)
	}
	f.now = f.now.Add(2 * time.Second)
	replacement, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok || replacement.Token == old.Token || replacement.Attempts != 2 {
		t.Fatal("replacement lease", ok, err)
	}
	if f.store.Reserve(ctx, old) == nil || f.store.Heartbeat(ctx, old, jobs.DefaultPolicy()) == nil {
		t.Fatal("old lease reused")
	}
	completion := aiJobCompletion(t)
	if _, err = f.store.Complete(ctx, old, completion); err == nil {
		t.Fatal("old lease published")
	}
	if _, err = f.store.Complete(ctx, replacement, completion); err != jobs.ErrBudget {
		t.Fatal("old reservation authorized new lease", err)
	}
	if err = f.store.Reserve(ctx, replacement); err != nil {
		t.Fatal(err)
	}
	if _, err = f.store.Complete(ctx, replacement, completion); err != nil {
		t.Fatal("replacement completion", err)
	}
}

func TestAIJobRepeatedInterruptionsStopAtAttemptBound(t *testing.T) {
	for _, reserve := range []bool{false, true} {
		f := newAIJobFixture(t)
		ctx := context.Background()
		f.input.AssetIDs = []string{selectionA}
		f.input.MaxCalls = 3
		job, err := f.store.Submit(ctx, f.input)
		if err != nil {
			t.Fatal(err)
		}
		for attempt := 1; attempt <= 3; attempt++ {
			lease, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
			if err != nil || !ok || lease.Attempts != attempt {
				t.Fatal("wrong attempt", attempt, err)
			}
			if reserve {
				if err = f.store.Reserve(ctx, lease); err != nil {
					t.Fatal(err)
				}
			}
			f.now = lease.ExpiresAt
			if n, err := f.store.Recover(ctx); err != nil || n != 1 {
				t.Fatal("recovery", n, err)
			}
			f.now = f.now.Add(2 * time.Second)
		}
		if _, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy()); err != nil || ok {
			t.Fatal("unbounded claims", ok, err)
		}
		got, err := f.store.Get(ctx, testUserID, job.ID)
		if err != nil || got.Items[0].State != "failed" || got.Items[0].Attempts != 3 {
			t.Fatal("not terminal", err)
		}
		expected := 0
		if reserve {
			expected = 3
		}
		if got.Calls != expected || got.Items[0].Calls != expected {
			t.Fatal("invented/refunded calls", got.Calls)
		}
	}
}
