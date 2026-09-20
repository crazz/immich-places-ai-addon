package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/jobs"
)

func TestAIExecutionReadinessRequiresPolicyAndObservedSupport(t *testing.T) {
	for _, state := range []string{"no-policy", "revision", "model", "capability", "disabled", "anonymous"} {
		t.Run(state, func(t *testing.T) {
			f := newAIVisualFixture(t)
			policies := attestedExecutionPolicy(t, f)
			if state == "no-policy" {
				policies = jobs.ExecutionPolicies{}
			}
			if state == "anonymous" {
				selectionSQL(t, f.image.db, "UPDATE ai_provider_versions SET secretCiphertext=NULL")
			}
			profiles, err := f.image.db.listAIProviders(context.Background(), testUserID, capabilities.ApplicabilityContext{AIEnabled: true, CurrentPolicyFingerprint: policyFingerprint(f.analyzer.dispatcher.Policy)})
			if err != nil || len(profiles) != 1 {
				t.Fatal(err)
			}
			p := profiles[0]
			want := "policy_required"
			switch state {
			case "revision":
				p.Revision++
			case "model":
				p.Model = "other-model"
			case "capability":
				p.CapabilityReport = nil
				want = "capability_required"
			case "disabled":
				p.Enabled = false
				want = "unavailable"
			case "anonymous", "no-policy":
				want = "ready"
			}
			h := &aiProviderHandlers{db: f.image.db, enabled: true, policy: f.analyzer.dispatcher.Policy, executionPolicies: policies}
			r, err := h.executionReadiness(context.Background(), testUserID, p)
			if err != nil || r == nil || r.Status != want {
				t.Fatal("wrong readiness authority", r, err, want)
			}
			if f.hits.Load() != 0 {
				t.Fatal("readiness inferred support using a probe")
			}
		})
	}
}
