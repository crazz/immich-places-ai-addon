package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionReservationsAtomicallyConsumeTokensAndCalls(t *testing.T) {
	f, p, req := productionFixture(t)
	ctx := context.Background()
	req.Configuration.Limits.MaxCalls = 3
	req.Consent.Configuration = req.Configuration
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	s := &aiProductionWorkerStore{aiJobStore: p.store, production: p}
	lease, claimed, err := s.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !claimed {
		t.Fatal(err)
	}
	start := make(chan struct{})
	answers := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() { <-start; answers <- s.Reserve(ctx, lease) }()
	}
	close(start)
	successes := 0
	for i := 0; i < 2; i++ {
		err := <-answers
		if err == nil {
			successes++
		} else if err != jobs.ErrBudget {
			t.Fatal(err)
		}
	}
	if successes != 1 {
		t.Fatal("more than one attested allowance reserved", successes)
	}
	var calls int
	var tokens int64
	if err := f.image.db.db.QueryRow("SELECT calls FROM ai_jobs WHERE id=?", job.ID).Scan(&calls); err != nil || calls != 1 {
		t.Fatal("call reservation", calls, err)
	}
	if err := f.image.db.db.QueryRow("SELECT sum(inputReserved+outputReserved) FROM ai_job_usage WHERE jobID=?", job.ID).Scan(&tokens); err != nil || tokens != 104000 {
		t.Fatal("token reservation missing", tokens, err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Usage.ReservedTokens != 104000 || progress.Usage.ReportedStatus != "unknown" {
		t.Fatal("usage incorrectly reported", progress.Usage, err)
	}
}
