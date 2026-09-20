package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIJobCancellationRejectsLateResultsAndPreservesCompleted(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	f.input.AssetIDs = append(f.input.AssetIDs, "33333333-3333-4333-8333-333333333333")
	f.input.MaxCalls = 9
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	completed, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok {
		t.Fatal(err)
	}
	if err = f.store.Reserve(ctx, completed); err != nil {
		t.Fatal(err)
	}
	resultID, err := f.store.Complete(ctx, completed, aiJobCompletion(t))
	if err != nil {
		t.Fatal(err)
	}
	running, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok {
		t.Fatal(err)
	}
	if err = f.store.Reserve(ctx, running); err != nil {
		t.Fatal(err)
	}
	if err = f.store.Cancel(ctx, "foreign", job.ID); err != jobs.ErrDenied {
		t.Fatal("foreign cancellation", err)
	}
	if err = f.store.Cancel(ctx, testUserID, job.ID); err != nil {
		t.Fatal(err)
	}
	if f.store.Authorize(ctx, running) == nil || f.store.Reserve(ctx, running) == nil || f.store.Heartbeat(ctx, running, jobs.DefaultPolicy()) == nil {
		t.Fatal("canceled authority admitted")
	}
	if _, err = f.store.Complete(ctx, running, aiJobCompletion(t)); err == nil {
		t.Fatal("late response published")
	}
	if _, ok, err = f.store.Claim(ctx, jobs.DefaultPolicy()); err != nil || ok {
		t.Fatal("canceled work claimed", err)
	}
	before, err := f.store.Get(ctx, testUserID, job.ID)
	if err != nil || !before.Canceled || before.Items[0].State != "succeeded" || before.Items[1].State != "running" || before.Items[2].State != "canceled" {
		t.Fatal("wrong cancellation states", before.Items, err)
	}
	f.now = running.ExpiresAt
	if n, err := f.store.Recover(ctx); err != nil || n != 1 {
		t.Fatal("canceled recovery", n, err)
	}
	after, err := f.store.Get(ctx, testUserID, job.ID)
	if err != nil || after.Items[1].State != "canceled" {
		t.Fatal("running cancel not reconciled", err)
	}
	if _, err = f.store.ReadAnalysis(ctx, testUserID, resultID); err != nil {
		t.Fatal("completed history lost", err)
	}
	var count int
	if err = f.db.db.QueryRow("SELECT count(*) FROM ai_analyses").Scan(&count); err != nil || count != 1 {
		t.Fatal("late analysis exists", count, err)
	}
}
