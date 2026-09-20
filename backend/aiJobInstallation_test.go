package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIJobInstallationRotationFencesUnfinishedWorkAtomically(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	f.input.AssetIDs = append(f.input.AssetIDs, "33333333-3333-4333-8333-333333333333")
	f.input.MaxCalls = 9
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
	id, err := f.store.Complete(ctx, first, aiJobCompletion(t))
	if err != nil {
		t.Fatal(err)
	}
	active, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok {
		t.Fatal(err)
	}
	if err = f.selection.bind(ctx, "https://replacement.example/api", "epoch-2"); err != nil {
		t.Fatal(err)
	}
	if f.selection.binding == f.input.Installation {
		t.Fatal("identity not rotated")
	}
	if f.store.Reserve(ctx, active) == nil {
		t.Fatal("old installation reserved")
	}
	if _, err = f.store.Complete(ctx, active, aiJobCompletion(t)); err == nil {
		t.Fatal("old installation published")
	}
	var unfinished, retained int
	if err = f.db.db.QueryRow("SELECT count(*) FROM ai_job_items WHERE jobID=? AND state NOT IN ('succeeded','canceled')", job.ID).Scan(&unfinished); err != nil || unfinished != 0 {
		t.Fatal("old work not canceled", unfinished, err)
	}
	if err = f.db.db.QueryRow("SELECT count(*) FROM ai_analyses WHERE id=?", id).Scan(&retained); err != nil || retained != 1 {
		t.Fatal("history not retained", retained, err)
	}
	current := newAIJobStore(f.db, f.selection.binding, true, f.store.now)
	if _, err = current.Get(ctx, testUserID, job.ID); err != jobs.ErrDenied {
		t.Fatal("old binding exposed as current", err)
	}
}

func TestAIJobFailedInstallationRotationRollsBackInvalidation(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	lease, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok {
		t.Fatal(err)
	}
	selectionSQL(t, f.db, `CREATE TRIGGER reject_installation BEFORE UPDATE ON ai_installation_identity BEGIN SELECT RAISE(ABORT,'rotation failed'); END`)
	if err = f.selection.bind(ctx, "https://replacement.example/api", "failed-epoch"); err == nil {
		t.Fatal("rotation failure ignored")
	}
	if f.selection.binding != f.input.Installation {
		t.Fatal("in-memory identity changed")
	}
	kept, err := f.store.Get(ctx, testUserID, job.ID)
	if err != nil || kept.Canceled || kept.Items[0].State != "running" || kept.Items[1].State != "queued" {
		t.Fatal("invalidation escaped rollback", kept.Items, err)
	}
	if err = f.store.Reserve(ctx, lease); err != nil {
		t.Fatal("old valid lease lost", err)
	}
}
