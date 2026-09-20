package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionRunsWithApplicationDefaultsWithoutOperatorPolicy(t *testing.T) {
	f, p, req := productionFixture(t)
	p.policies = jobs.ExecutionPolicies{}
	ctx := context.Background()
	profiles, err := f.image.db.listAIProviders(ctx, testUserID, capabilities.ApplicabilityContext{AIEnabled: true, CurrentPolicyFingerprint: p.fingerprint})
	if err != nil || len(profiles) != 1 {
		t.Fatal("profile unavailable", err)
	}
	h := &aiProviderHandlers{db: f.image.db, enabled: true, policy: f.analyzer.dispatcher.Policy}
	ready, err := h.executionReadiness(ctx, testUserID, profiles[0])
	if err != nil || ready.Status != "ready" || ready.Source != "application-defaults" || !ready.ContextAllowed || ready.CostStatus != "unknown" {
		t.Fatal("ordinary launch requires manual policy", ready, err)
	}
	req.Configuration.PolicyID = ready.PolicyID
	req.Configuration.Limits.MaxTokens = ready.MaxInputTokens + req.Configuration.Limits.OutputTokens
	req.Consent.Configuration = req.Configuration
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal("default admission failed", err)
	}
	worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
	if worked, err := worker.RunOne(ctx, make(chan time.Time)); err != nil || !worked {
		t.Fatal("default execution failed", worked, err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Counts["succeeded"] != 1 || progress.Items[0].ResultID == nil || progress.Usage.Calls != 1 || progress.Usage.CostStatus != "unknown" {
		t.Fatal("default result missing", progress, err)
	}
	_, writes := f.image.counts()
	if f.hits.Load() != 1 || writes != 0 {
		t.Fatal("unexpected external work", f.hits.Load(), writes)
	}
}

func TestAIProductionDefaultsRetainUsageAboveRequestedTokensWithoutBlocking(t *testing.T) {
	f, p, req := productionFixture(t)
	p.policies = jobs.ExecutionPolicies{}
	policy, id, ok := p.policies.Resolve(jobs.ExecutionBinding{Owner: testUserID, Installation: p.store.binding, Profile: req.Configuration.ProfileID, Revision: 1, Model: "bound-model", EgressFingerprint: p.fingerprint})
	if !ok {
		t.Fatal("defaults unavailable")
	}
	req.Configuration.PolicyID = id
	req.Configuration.Limits.MaxTokens = policy.MaxInputTokens + req.Configuration.Limits.OutputTokens
	req.Consent.Configuration = req.Configuration
	content, err := os.ReadFile("internal/ai/results/testdata/ai-analysis-result.unknown.json")
	if err != nil {
		t.Fatal(err)
	}
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": string(content)}}},
			"usage":   map[string]int{"prompt_tokens": 100001, "completion_tokens": 5000, "total_tokens": 105001},
		})
		return true
	}
	ctx := context.Background()
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
	if _, err := worker.RunOne(ctx, make(chan time.Time)); err != nil {
		t.Fatal(err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Blocked || progress.Counts["succeeded"] != 1 || progress.Usage.TotalReported == nil || *progress.Usage.TotalReported != 105001 || progress.Usage.CostStatus != "unknown" {
		t.Fatal("requested tokens treated as an attested provider cap", progress, err)
	}
	var violations int
	if err := f.image.db.db.QueryRow("SELECT count(*) FROM ai_execution_policy_violations").Scan(&violations); err != nil || violations != 0 {
		t.Fatal("automatic defaults fabricated an attestation violation", violations, err)
	}
	_, writes := f.image.counts()
	if f.hits.Load() != 1 || writes != 0 {
		t.Fatal("unexpected external work", f.hits.Load(), writes)
	}
}
