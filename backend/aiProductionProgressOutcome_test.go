package main

import (
	"context"
	"encoding/json"
	"immich-places-backend/internal/ai/jobs"
	"testing"
	"time"
)

func TestAIProductionProgressDistinguishesSuccessfulUnknownOutcome(t *testing.T) {
	f, p, req := productionFixture(t)
	ctx := context.Background()
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
	if _, err = worker.RunOne(ctx, make(chan time.Time)); err != nil {
		t.Fatal(err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(progress)
	var result struct {
		Items []struct{ State, Outcome string }
	}
	_ = json.Unmarshal(raw, &result)
	if result.Items[0].State != "succeeded" || result.Items[0].Outcome != "unknown" {
		t.Fatal("unknown outcome hidden behind execution state")
	}
}
