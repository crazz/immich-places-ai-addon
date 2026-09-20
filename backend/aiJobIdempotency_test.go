package main

import (
	"context"
	"sync"
	"testing"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIJobConcurrentIdenticalSubmissionsConverge(t *testing.T) {
	f := newAIJobFixture(t)
	const count = 12
	ids := make(chan string, count)
	var wg sync.WaitGroup
	for range count {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store := newAIJobStore(f.db, f.selection.binding, true, f.store.now)
			job, err := store.Submit(context.Background(), f.input)
			if err != nil {
				t.Error(err)
				return
			}
			ids <- job.ID
		}()
	}
	wg.Wait()
	close(ids)
	var first string
	for id := range ids {
		if first == "" {
			first = id
		}
		if first != id {
			t.Error("duplicate jobs")
		}
	}
	var jobs, items int
	if err := f.db.db.QueryRow("SELECT count(*) FROM ai_jobs").Scan(&jobs); err != nil {
		t.Fatal(err)
	}
	if err := f.db.db.QueryRow("SELECT count(*) FROM ai_job_items").Scan(&items); err != nil {
		t.Fatal(err)
	}
	if first == "" || jobs != 1 || items != 2 {
		t.Fatal("duplicate membership", jobs, items)
	}
}

func TestAIJobChangedRequestConflictsWithoutReplacingOriginal(t *testing.T) {
	for _, change := range []string{"membership", "language", "revision", "budget"} {
		t.Run(change, func(t *testing.T) {
			f := newAIJobFixture(t)
			ctx := context.Background()
			first, err := f.store.Submit(ctx, f.input)
			if err != nil {
				t.Fatal(err)
			}
			switch change {
			case "membership":
				f.input.AssetIDs = []string{selectionB}
				f.input.MaxCalls = 3
			case "language":
				f.input.Languages = []string{"uk"}
				f.input.PrimaryLanguage = "uk"
			case "revision":
				if _, err := f.db.updateAIProvider(ctx, testUserID, "profile", 1, providerInput()); err != nil {
					t.Fatal(err)
				}
				f.input.Revision = 2
			case "budget":
				f.input.MaxCalls = 5
			}
			if _, err := f.store.Submit(ctx, f.input); err != jobs.ErrConflict {
				t.Fatal("changed key did not conflict", err)
			}
			kept, err := f.store.Get(ctx, testUserID, first.ID)
			if err != nil || kept.Input.MaxCalls != 6 || kept.Input.Revision != 1 || len(kept.Items) != 2 || kept.Input.PrimaryLanguage != "en" {
				t.Fatal("original changed", err)
			}
		})
	}
}
