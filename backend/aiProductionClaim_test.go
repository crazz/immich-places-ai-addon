package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionConsumerCannotPromoteInternalJobs(t *testing.T) {
	f, p, req := productionFixture(t)
	ctx := context.Background()
	var digest string
	if err := f.image.db.db.QueryRow("SELECT digest FROM ai_selection_snapshots WHERE id=?", req.Configuration.SelectionToken).Scan(&digest); err != nil {
		t.Fatal(err)
	}
	_, err := p.store.Submit(ctx, jobs.Submission{Owner: testUserID, Installation: p.store.binding, Key: "old-internal", Profile: req.Configuration.ProfileID, Revision: 1, AssetIDs: []string{selectionA}, Languages: []string{"en"}, PrimaryLanguage: "en", SelectionDigest: digest, ConsentVersion: "visual-v1", MaxCalls: 1})
	if err != nil {
		t.Fatal(err)
	}
	workerStore := &aiProductionWorkerStore{aiJobStore: p.store, production: p}
	if _, claimed, err := workerStore.Claim(ctx, jobs.DefaultPolicy()); err != nil || claimed {
		t.Fatal("internal work promoted", claimed, err)
	}
	admitted, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	lease, claimed, err := workerStore.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !claimed || lease.JobID != admitted.ID {
		t.Fatal("production work unavailable", lease, claimed, err)
	}
}
