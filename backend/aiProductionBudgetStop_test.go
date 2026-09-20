package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/selection"
)

func TestAIProductionExhaustedTokenAllowanceStopsPendingImages(t *testing.T) {
	f, p, req := productionFixture(t)
	ctx := context.Background()
	seedAsset(t, f.image.db, selectionB, nil, nil, "2026-09-20")
	snapshot, err := p.selections.preview(ctx, selection.Input{Mode: "explicit", AssetIDs: []string{selectionA, selectionB}, Scope: &selection.Scope{View: "all"}}, testUserID)
	if err != nil {
		t.Fatal(err)
	}
	req.Configuration.SelectionToken = *snapshot.SnapshotID
	req.Configuration.Limits.MaxCalls = 2
	req.Consent.Configuration = req.Configuration
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
	if _, err = worker.RunOne(ctx, make(chan time.Time)); err != nil {
		t.Fatal(err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Counts["succeeded"] != 1 || progress.Counts["failed"] != 1 || progress.Items[1].Failure != "budget" || progress.Items[1].Attempts != 0 {
		t.Fatal("exhausted budget left pending image work", progress, err)
	}
	if f.hits.Load() != 1 {
		t.Fatal("unexpected dispatch")
	}
}
