package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"immich-places-backend/internal/ai/contextual"
	"immich-places-backend/internal/ai/jobs"
)

func researchProductionFixture(t *testing.T, sourceURL ...string) (*aiVisualFixture, *aiProductionJobs, jobs.Admission) {
	t.Helper()
	f, p, req := productionFixture(t)
	p.policies = jobs.ExecutionPolicies{}
	policy, id, ok := p.policies.Resolve(jobs.ExecutionBinding{Owner: testUserID, Installation: p.store.binding, Profile: req.Configuration.ProfileID, Revision: 1, Model: "bound-model", EgressFingerprint: p.fingerprint})
	if !ok {
		t.Fatal("default policy unavailable")
	}
	req.Configuration.PolicyID = id
	req.Configuration.Limits.MaxTokens = policy.MaxInputTokens + req.Configuration.Limits.OutputTokens
	req.Configuration.Mode = "research"
	req.Configuration.Context = &jobs.ContextChoices{Version: "context-v1", Classes: []contextual.Class{contextual.Hint}, Hint: "Possibly near a bridge"}
	req.Consent.Configuration = req.Configuration
	content, err := os.ReadFile("internal/ai/results/testdata/ai-analysis-result.synthetic.json")
	if err != nil {
		t.Fatal(err)
	}
	var answer map[string]any
	if err := json.Unmarshal(content, &answer); err != nil {
		t.Fatal(err)
	}
	answer["schema_version"] = "2.0"
	answer["descriptions"] = answer["descriptions"].([]any)[:1]
	reference := "https://example.org/reference"
	if len(sourceURL) > 0 {
		reference = sourceURL[0]
	}
	answer["sources"] = []any{map[string]any{"id": "reference", "url": reference, "title": "Reference", "relevance": "Matching bridge outline."}}
	candidate := answer["candidates"].([]any)[0].(map[string]any)
	candidate["source_refs"] = []any{"reference"}
	location := candidate["camera_location"].(map[string]any)
	location["estimated_radius_m"], location["radius_basis"], location["granularity"] = 5000, "model_estimate", "city"
	content, err = json.Marshal(answer)
	if err != nil {
		t.Fatal(err)
	}
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			return true
		}
		if !strings.Contains(string(body["response_format"]), "research_analysis") || !strings.Contains(string(body["messages"]), "Possibly near a bridge") {
			t.Error("Research contract or hint missing")
		}
		if err := json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": string(content)}}}}); err != nil {
			t.Error(err)
		}
		return true
	}
	return f, p, req
}

func TestAIResearchPublishesCoarsePrivateHistoryWithOrdinaryDefaults(t *testing.T) {
	f, p, req := researchProductionFixture(t)
	ctx := context.Background()
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal("Research admission unavailable", err)
	}
	worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
	if worked, err := worker.RunOne(ctx, make(chan time.Time)); err != nil || !worked {
		t.Fatal("Research execution unavailable", worked, err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Counts["succeeded"] != 1 || progress.Items[0].ResultID == nil {
		t.Fatal("Research completion unavailable", progress, err)
	}
	s := &aiResultStore{jobs: p.store}
	detail, err := s.detail(ctx, testUserID, *progress.Items[0].ResultID, "", "")
	if err != nil || detail.Proposal == nil {
		t.Fatal("Research history unavailable", err)
	}
	if detail.Proposal.Candidates[0].CameraLocation.EstimatedRadiusM.String() != "5000" || len(detail.Proposal.Sources) != 1 || detail.Provenance.Context == nil || detail.Entry.Mode != "research" {
		t.Fatal("Research fields lost")
	}
	_, writes := f.image.counts()
	if f.hits.Load() != 1 || writes != 0 || progress.Usage.Calls != 1 {
		t.Fatal("unexpected provider work or write")
	}
}
