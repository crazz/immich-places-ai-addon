package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"
)

func TestAIExecutionProfileEditRequiresItsOwnAttestation(t *testing.T) {
	f := newAIVisualFixture(t)
	hash := sha256.Sum256([]byte("ai-session"))
	if err := f.image.db.createSession(context.Background(), hex.EncodeToString(hash[:]), testUserID, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	cfg := &Config{AIEnabled: true, AIPublicOrigin: aiTestOrigin, AIProviderEgressPolicy: f.analyzer.dispatcher.Policy, AIExecutionPolicies: attestedExecutionPolicy(t, f)}
	handler := newAIProviderHandler(f.image.db, cfg, f.analyzer.dispatcher)
	body, _ := json.Marshal(map[string]any{"name": "Edited", "baseURL": f.provider.URL + "/v1", "model": "bound-model", "enabled": true, "expectedRevision": 1})
	response := aiRequest(handler, "PUT", "/ai/providers/"+f.request.ProfileID, string(body), aiTestOrigin, true)
	var profile aiProviderProfile
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &profile) != nil {
		t.Fatal("profile edit failed", response.Code)
	}
	if profile.Revision != 2 || profile.ExecutionReadiness == nil || profile.ExecutionReadiness.Status != "policy_required" || profile.ExecutionReadiness.PolicyID != "" {
		t.Fatal("profile edit reused or omitted readiness")
	}
	if f.hits.Load() != 0 {
		t.Fatal("profile edit probed provider")
	}
}
