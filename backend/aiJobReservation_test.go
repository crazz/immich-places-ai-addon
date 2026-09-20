package main

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIJobReservationsCannotExceedItemOrJobBudget(t *testing.T) {
	for _, cap := range []int{2, 6} {
		t.Run(string(rune('0'+cap)), func(t *testing.T) {
			f := newAIJobFixture(t)
			f.input.MaxCalls = cap
			ctx := context.Background()
			job, err := f.store.Submit(ctx, f.input)
			if err != nil {
				t.Fatal(err)
			}
			lease, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
			if err != nil || !ok {
				t.Fatal("claim", err)
			}
			var accepted atomic.Int32
			var wg sync.WaitGroup
			for range 12 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					err := f.store.Reserve(ctx, lease)
					if err == nil {
						accepted.Add(1)
					} else if err != jobs.ErrBudget {
						t.Error("wrong reservation failure", err)
					}
				}()
			}
			wg.Wait()
			want := min(cap, 3)
			loaded, err := f.store.Get(ctx, testUserID, job.ID)
			if err != nil {
				t.Fatal(err)
			}
			if accepted.Load() != int32(want) || loaded.Calls != want || loaded.Items[0].Calls != want || loaded.Items[1].Calls != 0 {
				t.Fatal("incorrect durable reservation", accepted.Load(), loaded.Calls, loaded.Items)
			}
		})
	}
}

func TestAIJobExhaustedBudgetTerminalizesUndispatchedItems(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	f.input.MaxCalls = 1
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
	if _, err = f.store.Complete(ctx, lease, aiJobCompletion(t)); err != nil {
		t.Fatal(err)
	}
	if _, ok, err = f.store.Claim(ctx, jobs.DefaultPolicy()); err != nil || ok {
		t.Fatal("over-budget claim", err)
	}
	got, err := f.store.Get(ctx, testUserID, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Items[0].State != "succeeded" || got.Items[1].State != "failed" || got.Items[1].Failure != "budget" || got.Items[1].Attempts != 0 || got.Calls != 1 {
		t.Fatal("budget left permanently queued work", got.Items)
	}
}
