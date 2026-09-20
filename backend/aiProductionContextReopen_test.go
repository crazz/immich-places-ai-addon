package main

import (
	"context"
	"immich-places-backend/internal/ai/jobs"
	"path/filepath"
	"testing"
	"time"
)

func TestAIProductionContextResultSurvivesReopen(t *testing.T) {
	f, p, req := contextProductionFixture(t)
	ctx := context.Background()
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
	if _, err = worker.RunOne(ctx, make(chan time.Time)); err != nil {
		t.Fatal(err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Items[0].ResultID == nil {
		t.Fatal("missing completion", err)
	}
	var seq int
	var name, path string
	if err = f.image.db.db.QueryRow("PRAGMA database_list").Scan(&seq, &name, &path); err != nil {
		t.Fatal(err)
	}
	f.image.db.close()
	db, err := newDatabase(filepath.Dir(path), "test-encryption-key")
	if err != nil {
		t.Fatal(err)
	}
	defer db.close()
	record, err := newAIJobStore(db, p.store.binding, true, time.Now).ReadAnalysis(ctx, testUserID, *progress.Items[0].ResultID)
	if err != nil || record.Metadata.Context == nil || record.Metadata.Context.Digest == "" || string(record.Metadata.Mode) != "context-assisted" || record.Metadata.Context.Consent.Version != "context-v1" {
		t.Fatal("context history lost across reopen", err)
	}
}
