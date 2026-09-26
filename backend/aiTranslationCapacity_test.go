package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"immich-places-backend/internal/ai/jobs"
)

func TestAITranslationSharesProviderCapacityWithAnalysis(t *testing.T) {
	f, s, req := translationFixture(t)
	entered, release := make(chan struct{}), make(chan struct{})
	dispatcher := translationProvider(t, f, s, func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": `{"language":"en","status":"complete","text":"A bridge."}`}}}})
	})
	req.ProfileRevision = 2
	req.Languages = []string{"en"}
	ctx := context.Background()
	if _, err := s.submit(ctx, testUserID, req); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := s.runOne(ctx, dispatcher); done <- err }()
	<-entered
	policy := jobs.DefaultPolicy()
	policy.Global = 1
	_, claimed, err := f.store.Claim(ctx, policy)
	close(release)
	if finished := <-done; finished != nil {
		t.Fatal(finished)
	}
	if err != nil || claimed {
		t.Fatal("analysis bypassed translation capacity", claimed, err)
	}
}

func TestAITranslationReservesTokensAndCostBeforeSending(t *testing.T) {
	for _, limit := range []string{"tokens", "cost", "request bytes"} {
		t.Run(limit, func(t *testing.T) {
			f, s, req := translationFixture(t)
			dispatcher := translationProvider(t, f, s, func(w http.ResponseWriter, r *http.Request) {
				t.Error("budget bypassed")
				http.Error(w, "blocked", 500)
			})
			req.ProfileRevision = 2
			if limit == "tokens" {
				req.MaxTokens = 1
			}
			if limit == "cost" {
				cap := int64(1)
				req.MaxEstimatedMicros = &cap
			}
			if limit == "request bytes" {
				policy, _, ok := s.policies.Resolve(jobs.ExecutionBinding{Owner: testUserID, Installation: f.store.binding, Profile: "profile", Revision: 2, Model: providerInput().Model, EgressFingerprint: s.fingerprint})
				if !ok {
					t.Fatal("missing test policy")
				}
				policy.Version = "execution-v1"
				policy.EvidenceRef = "synthetic-test"
				policy.MaxRequestBytes = 1
				raw, _ := json.Marshal([]jobs.ExecutionPolicy{policy})
				var err error
				s.policies, err = jobs.ParseExecutionPolicies(string(raw))
				if err != nil {
					t.Fatal(err)
				}
			}
			ctx := context.Background()
			if _, err := s.submit(ctx, testUserID, req); err != nil {
				t.Fatal(err)
			}
			if _, err := s.runOne(ctx, dispatcher); err != nil {
				t.Fatal(err)
			}
			run, err := s.submit(ctx, testUserID, req)
			if err != nil || run.Items[0].State != "failed" || run.Items[0].Failure != "budget" {
				t.Fatal("budget outcome missing", run, err)
			}
		})
	}
}

func TestAITranslationWaitsForAnalysisCapacity(t *testing.T) {
	f, s, req := translationFixture(t)
	dispatcher := translationProvider(t, f, s, func(w http.ResponseWriter, r *http.Request) {
		t.Error("translation bypassed analysis capacity")
		http.Error(w, "unavailable", 500)
	})
	req.ProfileRevision = 2
	ctx := context.Background()
	if _, err := s.submit(ctx, testUserID, req); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy()); !ok || err != nil {
		t.Fatal("analysis fixture unavailable", ok, err)
	}
	if worked, err := s.runOne(ctx, dispatcher); worked || err != nil {
		t.Fatal("translation failed to wait", worked, err)
	}
	run, err := s.submit(ctx, testUserID, req)
	if err != nil || run.Items[0].State != "queued" {
		t.Fatal("waiting work consumed", run, err)
	}
}
