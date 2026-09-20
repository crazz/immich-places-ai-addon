package main

import (
	"context"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/aiadapters/providerhttp"
	"path/filepath"
	"testing"
	"time"
)

func TestAIProductionContextRetryReusesEvidenceAcrossRestart(t *testing.T) {
	f, p, req := contextProductionFixture(t)
	ctx := context.Background()
	req.Configuration.Limits.MaxCalls = 3
	req.Configuration.Limits.MaxTokens = 312000
	req.Consent.Configuration = req.Configuration
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	store := &aiProductionWorkerStore{aiJobStore: p.store, production: p}
	lease, claimed, err := store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !claimed {
		t.Fatal(err)
	}
	guard := jobs.Guard{Authorize: func(ctx context.Context) error { return store.Authorize(ctx, lease) }, Reserve: func(ctx context.Context) error { return store.Reserve(ctx, lease) }}
	if _, err = p.executor(f.analyzer)(ctx, lease, guard); err != nil {
		t.Fatal(err)
	}
	var frozen string
	if err = f.image.db.db.QueryRow("SELECT bundleJSON FROM ai_job_context WHERE jobID=?", job.ID).Scan(&frozen); err != nil {
		t.Fatal(err)
	}
	if err = store.Fail(ctx, lease, jobs.Failure{Code: jobs.Transient}, time.Second); err != nil {
		t.Fatal(err)
	}
	var sequence int
	var name, location string
	if err = f.image.db.db.QueryRow("PRAGMA database_list").Scan(&sequence, &name, &location); err != nil {
		t.Fatal(err)
	}
	binding, policy := p.store.binding, f.analyzer.dispatcher.Policy
	f.image.db.close()
	db, err := newDatabase(filepath.Dir(location), "test-encryption-key")
	if err != nil {
		t.Fatal(err)
	}
	defer db.close()
	selections := newAISelectionStore(db)
	selections.enabled, selections.binding = true, binding
	images, err := newAIImagePreparer(db, selections, f.image.server.URL)
	if err != nil {
		t.Fatal(err)
	}
	analyzer, err := newAIVisualAnalyzer(db, images, newAIProviderDispatcher(db, true, policy, providerhttp.New(providerhttp.Options{})))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().Add(time.Minute)
	restarted := &aiProductionJobs{store: newAIJobStore(db, binding, true, func() time.Time { return now }), selections: selections, policies: p.policies, fingerprint: p.fingerprint}
	store = &aiProductionWorkerStore{aiJobStore: restarted.store, production: restarted}
	lease, claimed, err = store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !claimed {
		t.Fatal("reopened retry", err)
	}
	completion, err := restarted.executor(analyzer)(ctx, lease, guard)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.Complete(ctx, lease, completion); err != nil {
		t.Fatal(err)
	}
	var retained string
	if err = db.db.QueryRow("SELECT bundleJSON FROM ai_job_context WHERE jobID=?", job.ID).Scan(&retained); err != nil {
		t.Fatal(err)
	}
	if retained != frozen || f.hits.Load() != 2 {
		t.Fatal("frozen context changed across restart")
	}
	progress, err := restarted.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Counts["succeeded"] != 1 {
		t.Fatal("reopened completion unavailable", err)
	}
}
