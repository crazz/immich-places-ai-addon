package main

import (
	"encoding/json"
	"immich-places-backend/internal/ai/contextual"
	"immich-places-backend/internal/ai/jobs"
	"testing"
)

func contextProductionFixture(t *testing.T) (*aiVisualFixture, *aiProductionJobs, jobs.Admission) {
	t.Helper()
	f, p, req := productionFixture(t)
	policy, _, ok := p.policies.Find(jobs.ExecutionBinding{Owner: testUserID, Installation: p.store.binding, Profile: req.Configuration.ProfileID, Revision: 1, Model: "bound-model", EgressFingerprint: p.fingerprint})
	if !ok {
		t.Fatal("missing policy")
	}
	raw, _ := json.Marshal(policy)
	var cfg map[string]any
	_ = json.Unmarshal(raw, &cfg)
	cfg["context"] = true
	raw, _ = json.Marshal([]any{cfg})
	policies, err := jobs.ParseExecutionPolicies(string(raw))
	if err != nil {
		t.Fatal("Context policy unavailable", err)
	}
	p.policies = policies
	_, id, _ := p.policies.Find(policy.Binding)
	req.Configuration.PolicyID = id
	req.Configuration.Mode = "context-assisted"
	req.Configuration.Context = &jobs.ContextChoices{Version: "context-v1", Classes: []contextual.Class{}}
	req.Consent.Configuration = req.Configuration
	return f, p, req
}

func TestAIProductionContextPolicyRequiresExplicitAttestation(t *testing.T) {
	_, p, req := contextProductionFixture(t)
	policy, id, ok := p.policies.Find(jobs.ExecutionBinding{Owner: testUserID, Installation: p.store.binding, Profile: req.Configuration.ProfileID, Revision: 1, Model: "bound-model", EgressFingerprint: p.fingerprint})
	raw, _ := json.Marshal(policy)
	var cfg map[string]any
	_ = json.Unmarshal(raw, &cfg)
	if !ok || id != req.Configuration.PolicyID || cfg["context"] != true {
		t.Fatal("missing exact context attestation")
	}
}
