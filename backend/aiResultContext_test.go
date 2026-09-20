package main

import (
	"context"
	"immich-places-backend/internal/ai/jobs"
	"testing"
	"time"
)

func TestAIResultContextDetailRetainsFrozenEvidenceWithoutFreshSourceReads(t *testing.T) {
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
		t.Fatal(err)
	}
	p.store.enabled = false
	selectionSQL(t, f.image.db, `DELETE FROM assets WHERE userID=?`, testUserID)
	before, _ := f.image.counts()
	calls := f.hits.Load()
	s := &aiResultStore{jobs: p.store}
	detail, err := s.detail(ctx, testUserID, *progress.Items[0].ResultID, "", "")
	if err != nil || detail.Provenance == nil || detail.Provenance.Context == nil || detail.Entry.Mode != "context-assisted" || detail.Entry.SourceAvailable {
		t.Fatal("context history unavailable", err)
	}
	after, _ := f.image.counts()
	if before != after || calls != f.hits.Load() {
		t.Fatal("history refreshed source or dispatched")
	}
}
