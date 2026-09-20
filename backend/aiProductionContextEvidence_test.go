package main

import (
	"context"
	"encoding/json"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/results"
	"os"
	"testing"
)

func TestAIProductionCompletionRejectsInventedContextSources(t *testing.T) {
	f, p, req := contextProductionFixture(t)
	ctx := context.Background()
	_, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	s := &aiProductionWorkerStore{aiJobStore: p.store, production: p}
	lease, _, err := s.Claim(ctx, jobs.DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	completion, err := p.executor(f.analyzer)(ctx, lease, jobs.Guard{Authorize: func(ctx context.Context) error { return s.Authorize(ctx, lease) }, Reserve: func(ctx context.Context) error { return s.Reserve(ctx, lease) }})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("internal/ai/results/testdata/ai-analysis-result.synthetic.json")
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	_ = json.Unmarshal(raw, &payload)
	payload["descriptions"] = payload["descriptions"].([]any)[:1]
	payload["candidates"].([]any)[0].(map[string]any)["source_refs"] = []string{"invented"}
	raw, _ = json.Marshal(payload)
	validator, _ := results.New()
	completion.Proposal, err = validator.Validate(raw, results.Context{Mode: results.ContextAssisted, Completion: results.Complete, Languages: []string{"en"}, PrimaryLanguage: "en", Sources: []results.Source{{ID: "invented"}}})
	if err != nil {
		t.Fatal("test proposal invalid", err)
	}
	if _, err = s.Complete(ctx, lease, completion); err == nil {
		t.Fatal("model source set replaced stored authority")
	}
	var count int
	_ = f.image.db.db.QueryRow("SELECT count(*) FROM ai_analyses").Scan(&count)
	if count != 0 {
		t.Fatal("invented evidence published")
	}
}
