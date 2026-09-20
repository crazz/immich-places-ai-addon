package main

import (
	"context"
	"encoding/json"
	"immich-places-backend/internal/ai/contextual"
	"immich-places-backend/internal/ai/jobs"
	"net/http"
	"testing"
	"time"
)

func TestAIProductionContextRetryNeverReplacesChangedEvidence(t *testing.T) {
	f, p, req := contextProductionFixture(t)
	ctx := context.Background()
	req.Configuration.Context = &jobs.ContextChoices{Version: "context-v1", Classes: []contextual.Class{contextual.Capture}}
	req.Configuration.Limits.MaxCalls = 3
	req.Configuration.Limits.MaxTokens = 312000
	req.Consent.Configuration = req.Configuration
	capture := "2026-09-20T12:00:00Z"
	f.image.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path != "/api/assets/"+selectionA {
			return false
		}
		m := f.image.metadata()
		m["exifInfo"] = map[string]any{"dateTimeOriginal": capture}
		_ = json.NewEncoder(w).Encode(m)
		return true
	}
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	s := &aiProductionWorkerStore{aiJobStore: p.store, production: p}
	lease, _, err := s.Claim(ctx, jobs.DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	guard := jobs.Guard{Authorize: func(ctx context.Context) error { return s.Authorize(ctx, lease) }, Reserve: func(ctx context.Context) error { return s.Reserve(ctx, lease) }}
	if _, err = p.executor(f.analyzer)(ctx, lease, guard); err != nil {
		t.Fatal(err)
	}
	var frozen string
	if err = f.image.db.db.QueryRow("SELECT bundleJSON FROM ai_job_context WHERE jobID=?", job.ID).Scan(&frozen); err != nil {
		t.Fatal(err)
	}
	if err = s.Fail(ctx, lease, jobs.Failure{Code: jobs.Transient}, time.Second); err != nil {
		t.Fatal(err)
	}
	now := time.Now().Add(time.Minute)
	p.store.now = func() time.Time { return now }
	capture = "2026-09-21T12:00:00Z"
	lease, claimed, err := s.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !claimed {
		t.Fatal("retry claim", err)
	}
	if _, err = p.executor(f.analyzer)(ctx, lease, guard); err == nil {
		t.Fatal("changed evidence accepted")
	}
	var retained string
	_ = f.image.db.db.QueryRow("SELECT bundleJSON FROM ai_job_context WHERE jobID=?", job.ID).Scan(&retained)
	if retained != frozen || f.hits.Load() != 1 {
		t.Fatal("changed evidence replaced or dispatched")
	}
}
