package main

import (
	"context"
	"encoding/json"
	"immich-places-backend/internal/ai/contextual"
	"immich-places-backend/internal/ai/jobs"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestAIProductionChangedContextFailsOnlyItsItem(t *testing.T) {
	f, p, req := contextProductionFixture(t)
	ctx := context.Background()
	req.Configuration.Context = &jobs.ContextChoices{Version: "context-v1", Classes: []contextual.Class{contextual.Capture}}
	req.Consent.Configuration = req.Configuration
	var capture atomic.Value
	capture.Store("2026-09-20T12:00:00Z")
	f.image.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path != "/api/assets/"+selectionA {
			return false
		}
		m := f.image.metadata()
		m["exifInfo"] = map[string]any{"dateTimeOriginal": capture.Load().(string)}
		_ = json.NewEncoder(w).Encode(m)
		return true
	}
	f.handle = func(http.ResponseWriter, *http.Request) bool { capture.Store("2026-09-21T12:00:00Z"); return false }
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
	if _, err = worker.RunOne(ctx, make(chan time.Time)); err != nil {
		t.Fatal(err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Blocked || progress.Counts["failed"] != 1 || progress.Items[0].ResultID != nil || f.hits.Load() != 1 {
		t.Fatal("changed evidence did not remain an item failure", progress, err)
	}
}
