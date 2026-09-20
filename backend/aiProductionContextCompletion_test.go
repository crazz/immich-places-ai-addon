package main

import (
	"context"
	"encoding/json"
	"immich-places-backend/internal/ai/contextual"
	"immich-places-backend/internal/ai/jobs"
	"testing"
	"time"
)

func TestAIProductionContextCompletionRetainsTrustedEvidence(t *testing.T) {
	f, p, req := contextProductionFixture(t)
	ctx := context.Background()
	req.Configuration.Context = &jobs.ContextChoices{Version: "context-v1", Classes: []contextual.Class{contextual.Hint}, Hint: "near a bridge"}
	req.Consent.Configuration = req.Configuration
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	s := &aiProductionWorkerStore{aiJobStore: p.store, production: p}
	worker := jobs.Worker{Store: s, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
	if worked, err := worker.RunOne(ctx, make(chan time.Time)); err != nil || !worked {
		t.Fatal("worker", err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Counts["succeeded"] != 1 {
		t.Fatal("context completion unavailable", progress, err)
	}
	record, err := p.store.ReadAnalysis(ctx, testUserID, *progress.Items[0].ResultID)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(record.Metadata)
	var meta map[string]json.RawMessage
	_ = json.Unmarshal(raw, &meta)
	if string(meta["Mode"]) != "\"context-assisted\"" || len(meta["Context"]) == 0 {
		t.Fatal("trusted context provenance missing")
	}
	var retained contextual.Metadata
	if json.Unmarshal(meta["Context"], &retained) != nil || len(retained.Sources) != 1 || retained.Sources[0].ID != "hint" || retained.Digest == "" {
		t.Fatal("source set lost")
	}
}
