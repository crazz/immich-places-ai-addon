package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func completeAIJob(t *testing.T, f *aiJobFixture, key string) jobs.Job {
	t.Helper()
	ctx := context.Background()
	input := f.input
	input.Key = key
	job, err := f.store.Submit(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	lease, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok || lease.JobID != job.ID {
		t.Fatal("fixture claim", err)
	}
	if err = f.store.Reserve(ctx, lease); err != nil {
		t.Fatal(err)
	}
	if _, err = f.store.Complete(ctx, lease, aiJobCompletion(t)); err != nil {
		t.Fatal(err)
	}
	return job
}

func TestAIJobCleanupIsBoundedPrivateAndTerminalOnly(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	f.input.AssetIDs = []string{selectionA}
	f.input.MaxCalls = 3
	for _, key := range []string{"old-a", "old-b", "old-c"} {
		completeAIJob(t, f, key)
	}
	if err := f.db.createUser(ctx, "other-owner", "other-history@example.com", "hash"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.createAIProvider(ctx, "other-owner", "profile", providerInput()); err != nil {
		t.Fatal(err)
	}
	f.input.Owner = "other-owner"
	other := completeAIJob(t, f, "other-old")
	f.input.Owner = testUserID
	activeInput := f.input
	activeInput.Key = "active"
	active, err := f.store.Submit(ctx, activeInput)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy()); err != nil || !ok {
		t.Fatal(err)
	}
	f.now = f.now.Add(48 * time.Hour)
	recent := completeAIJob(t, f, "recent")
	cutoff := f.now.Add(-24 * time.Hour)
	if n, err := f.store.PurgeBefore(ctx, testUserID, cutoff, 1); err != nil || n != 1 {
		t.Fatal("bounded cleanup", n, err)
	}
	if n, err := f.store.PurgeBefore(ctx, testUserID, cutoff, 100); err != nil || n != 2 {
		t.Fatal("remaining cleanup", n, err)
	}
	for _, kept := range []jobs.Job{active, recent, other} {
		if _, err := f.store.Get(ctx, kept.Input.Owner, kept.ID); err != nil {
			t.Fatal("protected job deleted", err)
		}
	}
	var jobsCount, resultsCount int
	if err = f.db.db.QueryRow("SELECT count(*) FROM ai_jobs").Scan(&jobsCount); err != nil {
		t.Fatal(err)
	}
	if err = f.db.db.QueryRow("SELECT count(*) FROM ai_analyses").Scan(&resultsCount); err != nil {
		t.Fatal(err)
	}
	if jobsCount != 3 || resultsCount != 2 {
		t.Fatal("incorrect cleanup cascade", jobsCount, resultsCount)
	}
}

func TestAIJobCleanupRejectsInvalidBoundsAndWorksWhenAIDisabled(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	f.input.AssetIDs = []string{selectionA}
	f.input.MaxCalls = 3
	completeAIJob(t, f, "old")
	f.now = f.now.Add(time.Hour)
	cases := []struct {
		owner  string
		cutoff time.Time
		limit  int
	}{
		{"", f.now, 1}, {testUserID, time.Time{}, 1}, {testUserID, f.now.Add(time.Second), 1}, {testUserID, f.now, 0}, {testUserID, f.now, 101},
	}
	for _, tc := range cases {
		if n, err := f.store.PurgeBefore(ctx, tc.owner, tc.cutoff, tc.limit); err != jobs.ErrInvalid || n != 0 {
			t.Fatal("invalid cleanup admitted", n, err)
		}
	}
	f.store.enabled = false
	if n, err := f.store.PurgeBefore(ctx, testUserID, f.now, 1); err != nil || n != 1 {
		t.Fatal("disabled-AI cleanup blocked", n, err)
	}
}
