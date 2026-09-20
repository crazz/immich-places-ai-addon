package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionEstimatedCostCapIsConservativeAcrossRetries(t *testing.T) {
	f, p, req := productionFixture(t)
	ctx := context.Background()
	binding := jobs.ExecutionBinding{Owner: testUserID, Installation: p.store.binding, Profile: req.Configuration.ProfileID, Revision: 1, Model: "bound-model", EgressFingerprint: p.fingerprint}
	policy, _, _ := p.policies.Find(binding)
	input, output := int64(1000000), int64(2000000)
	policy.Currency = "USD"
	policy.InputMicrosPerMillion = &input
	policy.OutputMicrosPerMillion = &output
	raw, _ := json.Marshal([]jobs.ExecutionPolicy{policy})
	var err error
	p.policies, err = jobs.ParseExecutionPolicies(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	_, id, _ := p.policies.Find(binding)
	req.Configuration.PolicyID = id
	cap := int64(108000)
	req.Configuration.Limits = jobs.Limits{MaxCalls: 3, MaxTokens: 312000, OutputTokens: 4000, MaxEstimatedMicros: &cap}
	req.Consent.Configuration = req.Configuration
	now := time.Now()
	p.store.now = func() time.Time { return now }
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	s := &aiProductionWorkerStore{aiJobStore: p.store, production: p}
	lease, _, err := s.Claim(ctx, jobs.DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Reserve(ctx, lease); err != nil {
		t.Fatal(err)
	}
	if err = s.Fail(ctx, lease, jobs.Failure{Code: jobs.Transient}, 0); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Minute)
	lease, claimed, err := s.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !claimed {
		t.Fatal(err)
	}
	if err = s.Reserve(ctx, lease); err != jobs.ErrBudget {
		t.Fatal("estimated cost cap exceeded", err)
	}
	var cost int64
	if err = f.image.db.db.QueryRow("SELECT sum(estimatedMicros) FROM ai_job_usage WHERE jobID=?", job.ID).Scan(&cost); err != nil || cost != cap {
		t.Fatal("cost estimate missing", cost, err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Usage.CostStatus != "estimated" || progress.Usage.ReservedTokens != 104000 || progress.Usage.Calls != 1 {
		t.Fatal("cost/usage conflated", progress.Usage, err)
	}
}
