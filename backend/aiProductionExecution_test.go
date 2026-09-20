package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionCompletesRealVisualAfterSnapshotCleanup(t *testing.T) {
	f, p, req := productionFixture(t)
	ctx := context.Background()
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.image.db, "DELETE FROM ai_selection_snapshots")
	s := &aiProductionWorkerStore{aiJobStore: p.store, production: p}
	worker := jobs.Worker{Store: s, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
	worked, err := worker.RunOne(ctx, make(chan time.Time))
	if err != nil || !worked {
		t.Fatal("worker did not finish", worked, err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Counts["succeeded"] != 1 || progress.Items[0].ResultID == nil || progress.Usage.Calls != 1 || progress.Usage.ReservedTokens != 104000 {
		t.Fatal("result not atomically published", progress, err)
	}
	record, err := p.store.ReadAnalysis(ctx, testUserID, *progress.Items[0].ResultID)
	if err != nil || record.Outcome != "unknown" || record.Metadata.Profile != req.Configuration.ProfileID || record.Metadata.SourceDigest == "" || record.Metadata.SelectionDigest == "" {
		t.Fatal("canonical result/provenance missing", record, err)
	}
	_, bad := f.image.counts()
	if f.hits.Load() != 1 || bad != 0 {
		t.Fatal("unexpected upstream requests", f.hits.Load(), bad)
	}
}
