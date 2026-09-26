package main

import (
	"context"
	"encoding/json"
	"immich-places-backend/internal/ai/jobs"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestAITranslationAccountsUsageAndStopsAtRunBudget(t *testing.T) {
	f, s, req := translationFixture(t)
	var hits atomic.Int32
	dispatcher := translationProvider(t, f, s, func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": `{"language":"en","status":"complete","text":"A bridge."}`}}}, "usage": map[string]int{"prompt_tokens": 100, "completion_tokens": 10, "total_tokens": 110}})
	})
	req.ProfileRevision = 2
	req.MaxTokens = 104096
	ctx := context.Background()
	run, err := s.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err = s.runOne(ctx, dispatcher); err != nil {
			t.Fatal(err)
		}
	}
	if hits.Load() != 1 {
		t.Fatal("run token ceiling bypassed", hits.Load())
	}
	var reserved, reported int64
	if err = f.db.db.QueryRow("SELECT inputReserved+outputReserved,totalReported FROM ai_translation_usage WHERE runID=? AND language='en'", run.ID).Scan(&reserved, &reported); err != nil || reserved != 104096 || reported != 110 {
		t.Fatal("usage not retained", reserved, reported, err)
	}
}

func TestAITranslationStopsAfterExplicitPolicyUsageViolation(t *testing.T) {
	f, s, req := translationFixture(t)
	var hits atomic.Int32
	dispatcher := translationProvider(t, f, s, func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": `{"language":"en","status":"complete","text":"A bridge."}`}}}, "usage": map[string]int{"prompt_tokens": 101, "completion_tokens": 10, "total_tokens": 111}})
	})
	req.ProfileRevision = 2
	policy, _, _ := s.policies.Resolve(jobs.ExecutionBinding{Owner: testUserID, Installation: f.store.binding, Profile: "profile", Revision: 2, Model: providerInput().Model, EgressFingerprint: s.fingerprint})
	policy.Version = "execution-v1"
	policy.EvidenceRef = "synthetic"
	policy.MaxInputTokens = 100
	raw, _ := json.Marshal([]jobs.ExecutionPolicy{policy})
	var err error
	s.policies, err = jobs.ParseExecutionPolicies(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	run, err := s.submit(context.Background(), testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err = s.runOne(context.Background(), dispatcher); err != nil {
			t.Fatal(err)
		}
	}
	var violations int
	if err = f.db.db.QueryRow("SELECT count(*) FROM ai_execution_policy_violations WHERE policyID=?", run.PolicyID).Scan(&violations); err != nil || violations != 1 || hits.Load() != 1 {
		t.Fatal("explicit policy violation did not stop dispatch", violations, hits.Load(), err)
	}
	got, err := s.submit(context.Background(), testUserID, req)
	if err != nil || got.Items[0].Text != nil || got.Items[0].State != "interrupted" {
		t.Fatal("violating reply published", got, err)
	}
}
