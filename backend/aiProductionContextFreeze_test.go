package main

import (
	"context"
	"immich-places-backend/internal/ai/jobs"
	"testing"
)

func TestAIProductionFreezesContextBeforeDispatch(t *testing.T) {
	f, p, req := contextProductionFixture(t)
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
	completion, err := p.executor(f.analyzer)(ctx, lease, jobs.Guard{Authorize: func(ctx context.Context) error { return s.Authorize(ctx, lease) }, Reserve: func(ctx context.Context) error {
		var count int
		if e := f.image.db.db.QueryRow("SELECT count(*) FROM ai_job_context WHERE userID=? AND jobID=? AND itemID=?", testUserID, job.ID, lease.ItemID).Scan(&count); e != nil || count != 1 {
			t.Errorf("context was not durably frozen before reservation: %v", e)
			return jobs.ErrDenied
		}
		return s.Reserve(ctx, lease)
	}})
	if err != nil {
		t.Fatal("Context attempt unavailable", err)
	}
	if completion.PromptVersion != "context-assisted-v1" {
		t.Fatal("Visual prompt used for Context")
	}
	var raw string
	if err = f.image.db.db.QueryRow("SELECT bundleJSON FROM ai_job_context WHERE jobID=?", job.ID).Scan(&raw); err != nil || len(raw) == 0 {
		t.Fatal("frozen evidence missing", err)
	}
	if f.hits.Load() != 1 {
		t.Fatal("wrong call count")
	}
}
