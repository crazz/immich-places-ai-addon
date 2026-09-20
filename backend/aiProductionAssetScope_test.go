package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionAssetLossStaysLocalAndOwnerLossBlocks(t *testing.T) {
	for _, kind := range []string{"asset-after-resolution", "owner-before-preparation"} {
		t.Run(kind, func(t *testing.T) {
			f, p, req := productionFixture(t)
			ctx := context.Background()
			job, err := p.submit(ctx, testUserID, req)
			if err != nil {
				t.Fatal(err)
			}
			expected := "failed"
			if kind == "asset-after-resolution" {
				f.analyzer.dispatcher.AfterResolve = func(context.Context) { selectionSQL(t, f.image.db, "DELETE FROM assets WHERE immichID=?", selectionA) }
			} else {
				selectionSQL(t, f.image.db, "UPDATE users SET immichAPIKey=NULL WHERE ID=?", testUserID)
				expected = "blocked"
			}
			worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
			if _, err = worker.RunOne(ctx, make(chan time.Time)); err != nil {
				t.Fatal(err)
			}
			progress, err := p.progress(ctx, testUserID, job.ID)
			if err != nil || progress.Counts[expected] != 1 || progress.Blocked != (expected == "blocked") || f.hits.Load() != 0 {
				t.Fatal("failure affected wrong scope", progress, err)
			}
		})
	}
}
