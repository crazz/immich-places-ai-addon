package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func attestedExecutionPolicy(t *testing.T, f *aiVisualFixture) jobs.ExecutionPolicies {
	t.Helper()
	p := jobs.ExecutionPolicy{Version: "execution-v1", Binding: jobs.ExecutionBinding{Owner: testUserID, Installation: f.image.store.binding, Profile: f.request.ProfileID, Revision: f.request.Revision, Model: "bound-model", EgressFingerprint: policyFingerprint(f.analyzer.dispatcher.Policy)}, OutputField: "max_completion_tokens", MaxInputTokens: 100000, MaxOutputTokens: 4000, MaxRequestBytes: 15 << 20, MaxImageBytes: 10 << 20, EvidenceRef: "private-attestation-document"}
	raw, err := json.Marshal([]jobs.ExecutionPolicy{p})
	if err != nil {
		t.Fatal(err)
	}
	policies, err := jobs.ParseExecutionPolicies(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	return policies
}

func TestAIExecutionReadinessIsPrivateAttestationWithoutProviderCalls(t *testing.T) {
	f := newAIVisualFixture(t)
	hash := sha256.Sum256([]byte("ai-session"))
	if err := f.image.db.createSession(context.Background(), hex.EncodeToString(hash[:]), testUserID, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	cfg := &Config{AIEnabled: true, AIPublicOrigin: aiTestOrigin, AIProviderEgressPolicy: f.analyzer.dispatcher.Policy, AIExecutionPolicies: attestedExecutionPolicy(t, f)}
	handler := newAIProviderHandler(f.image.db, cfg, f.analyzer.dispatcher)
	before, _ := f.image.counts()
	response := aiRequest(handler, "GET", "/ai/providers", "", "", true)
	var body struct {
		Items []struct {
			ExecutionReadiness struct {
				Status, PolicyID, Source, CostStatus string
				MaxInputTokens, MaxOutputTokens      int64
			}
			CapabilityReport json.RawMessage
		}
	}
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &body) != nil || len(body.Items) != 1 {
		t.Fatal("profile readiness response failed", response.Code)
	}
	r := body.Items[0].ExecutionReadiness
	if r.Status != "ready" || r.Source != "operator-attested" || r.CostStatus != "unknown" || len(r.PolicyID) != 64 || r.MaxInputTokens != 100000 || r.MaxOutputTokens != 4000 {
		t.Fatal("attestation readiness missing or conflated", response.Body.String())
	}
	if len(body.Items[0].CapabilityReport) == 0 || strings.Contains(response.Body.String(), "private-attestation-document") || strings.Contains(response.Body.String(), "provider-secret") {
		t.Fatal("evidence missing or private policy leaked")
	}
	after, bad := f.image.counts()
	if before != after || bad != 0 || f.hits.Load() != 0 {
		t.Fatal("reading readiness performed external work")
	}
}
