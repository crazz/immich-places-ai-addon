package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionOverAllowanceInvalidatesPolicyEvenForInvalidOutput(t *testing.T) {
	f, p, req := productionFixture(t)
	ctx := context.Background()
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{}, "usage": map[string]int{"prompt_tokens": 100001, "completion_tokens": 50, "total_tokens": 100051}})
		return true
	}
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
	if _, err = worker.RunOne(ctx, make(chan time.Time)); err != nil {
		t.Fatal(err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || !progress.Blocked || progress.Counts["blocked"] != 1 || progress.Items[0].ResultID != nil {
		t.Fatal("violated policy not blocked", progress, err)
	}
	req.IdempotencyKey = "after-violation"
	if _, err = p.submit(ctx, testUserID, req); err != jobs.ErrDenied {
		t.Fatal("violated policy readmitted", err)
	}
	productionSession(t, f)
	h := newAIProviderHandler(f.image.db, &Config{AIEnabled: true, AIProviderEgressPolicy: f.analyzer.dispatcher.Policy, AIExecutionPolicies: p.policies}, f.analyzer.dispatcher)
	rec := aiRequest(h, "GET", "/ai/providers", "", "", true)
	var response struct {
		Items []struct{ ExecutionReadiness struct{ Status string } }
	}
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &response) != nil || len(response.Items) != 1 || response.Items[0].ExecutionReadiness.Status != "policy_violated" {
		t.Fatal("readiness hid violation", rec.Body.String())
	}
	if f.hits.Load() != 1 {
		t.Fatal("extra provider calls")
	}
}
