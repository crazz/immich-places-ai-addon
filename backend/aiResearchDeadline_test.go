package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIResearchCarriesOneTenMinuteDeadlineThroughDispatch(t *testing.T) {
	f, p, req := researchProductionFixture(t)
	ctx := context.Background()
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	var deadline time.Time
	f.analyzer.dispatcher.AfterResolve = func(ctx context.Context) {
		got, ok := ctx.Deadline()
		if !ok || !got.Equal(deadline) {
			t.Error("analysis/dispatch shortened or extended worker deadline")
		}
	}
	execute := p.executor(f.analyzer)
	worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: func(ctx context.Context, lease jobs.Lease, guard jobs.Guard) (jobs.Completion, error) {
		var ok bool
		deadline, ok = ctx.Deadline()
		if !ok || time.Until(deadline) <= 9*time.Minute || time.Until(deadline) > 10*time.Minute {
			t.Fatal("Research retained Visual timeout")
		}
		return execute(ctx, lease, guard)
	}}
	if _, err := worker.RunOne(ctx, make(chan time.Time)); err != nil {
		t.Fatal(err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Counts["succeeded"] != 1 {
		t.Fatal("Research failed", err)
	}
}

func TestAIResearchHonorsShorterInstallationDeadline(t *testing.T) {
	f, p, req := researchProductionFixture(t)
	if _, err := p.submit(context.Background(), testUserID, req); err != nil {
		t.Fatal(err)
	}
	policy := jobs.DefaultPolicy()
	policy.ResearchDuration = 3 * time.Minute
	execute := p.executor(f.analyzer)
	worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: policy, Execute: func(ctx context.Context, lease jobs.Lease, guard jobs.Guard) (jobs.Completion, error) {
		deadline, _ := ctx.Deadline()
		if remaining := time.Until(deadline); remaining > 3*time.Minute || remaining < 2*time.Minute {
			t.Fatal("installation timeout ignored", remaining)
		}
		return execute(ctx, lease, guard)
	}}
	if _, err := worker.RunOne(context.Background(), make(chan time.Time)); err != nil {
		t.Fatal(err)
	}
}
