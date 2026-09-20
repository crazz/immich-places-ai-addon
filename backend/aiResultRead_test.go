package main

import (
	"context"
	"immich-places-backend/internal/ai/review"
	"testing"
)

func TestAIResultReadSurvivesGPSRemovalReopenAndDisabledExecution(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	seedAsset(t, f.db, selectionA, nil, nil, "2026-09-20")
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.store.Cancel(ctx, testUserID, job.ID); err != nil {
		t.Fatal(err)
	}
	s := &aiResultStore{jobs: f.store}
	page, err := s.list(ctx, testUserID, review.Query{Limit: 30})
	if err != nil || len(page.Items) != 2 {
		t.Fatal("history unavailable", err)
	}
	selectionSQL(t, f.db, "UPDATE assets SET latitude=0,longitude=0 WHERE userID=? AND immichID=?", testUserID, selectionA)
	selectionSQL(t, f.db, "DELETE FROM assets WHERE userID=? AND immichID=?", testUserID, selectionA)
	f.reopen(t)
	f.store.enabled = false
	s.jobs = f.store
	page, err = s.list(ctx, testUserID, review.Query{Limit: 30})
	if err != nil || len(page.Items) != 2 {
		t.Fatal("retained history lost", err)
	}
	for _, item := range page.Items {
		if item.AnalysisID != nil || item.ProposalOutcome != nil || item.ExecutionState != "canceled" || item.ReviewState != "unreviewed" || item.WriteState != "not_requested" || item.SourceAvailable {
			t.Fatal("invented state or source", item)
		}
	}
}
