package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionWorkerBlocksAuthorityLostAtPublication(t *testing.T) {
	f, p, req := productionFixture(t)
	ctx := context.Background()
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	execute := p.executor(f.analyzer)
	worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: func(ctx context.Context, lease jobs.Lease, guard jobs.Guard) (jobs.Completion, error) {
		result, err := execute(ctx, lease, guard)
		if err != nil {
			return result, err
		}
		selectionSQL(t, f.image.db, "UPDATE ai_provider_profiles SET enabled=0")
		return result, nil
	}}
	if _, err = worker.RunOne(ctx, make(chan time.Time)); err != nil {
		t.Fatal("authority failure escaped without state transition", err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || !progress.Blocked || progress.Counts["blocked"] != 1 || progress.Items[0].ResultID != nil {
		t.Fatal("publication lost authority left running", progress, err)
	}
}
