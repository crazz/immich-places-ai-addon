package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionPublicationRechecksProviderAtCommit(t *testing.T) {
	f, p, req := productionFixture(t)
	ctx := context.Background()
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	s := &aiProductionWorkerStore{aiJobStore: p.store, production: p}
	lease, claimed, err := s.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !claimed {
		t.Fatal(err)
	}
	completion, err := p.executor(f.analyzer)(ctx, lease, jobs.Guard{Authorize: func(ctx context.Context) error { return s.Authorize(ctx, lease) }, Reserve: func(ctx context.Context) error { return s.Reserve(ctx, lease) }})
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.image.db, "UPDATE ai_provider_profiles SET enabled=0")
	if id, err := s.Complete(ctx, lease, completion); err != jobs.ErrDenied || id != "" {
		t.Fatal("stale provider completion committed", id, err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Items[0].ResultID != nil {
		t.Fatal("late result visible", progress, err)
	}
	if err := s.Fail(ctx, lease, jobs.Failure{Code: jobs.Blocked}, time.Duration(0)); err != nil {
		t.Fatal(err)
	}
}
