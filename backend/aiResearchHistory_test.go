package main

import (
	"context"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/review"
	"path/filepath"
	"testing"
	"time"
)

func TestAIResearchHistorySurvivesReopenSourceLossAndDisabledExecution(t *testing.T) {
	f, p, req := researchProductionFixture(t)
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
		t.Fatal("result missing", err)
	}
	resultID := *progress.Items[0].ResultID
	selectionSQL(t, f.image.db, "DELETE FROM assets WHERE userID=? AND immichID=?", testUserID, selectionA)
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
	store := newAIJobStore(db, p.store.binding, false, time.Now)
	history := &aiResultStore{jobs: store}
	detail, err := history.detail(ctx, testUserID, resultID, "", "")
	if err != nil || detail.Proposal == nil || detail.Proposal.SchemaVersion != "2.0" || len(detail.Proposal.Sources) != 1 || detail.Proposal.Candidates[0].CameraLocation.EstimatedRadiusM.String() != "5000" || detail.Entry.SourceAvailable {
		t.Fatal("Research history lost", err)
	}
	if _, err = history.detail(ctx, "foreign-owner", resultID, "", ""); err == nil {
		t.Fatal("foreign history disclosed")
	}
	page, err := history.list(ctx, testUserID, review.Query{Limit: 30})
	if err != nil || len(page.Items) != 1 {
		t.Fatal("history listing lost", err)
	}
	if f.hits.Load() != 1 {
		t.Fatal("history repeated inference")
	}
	selectionSQL(t, db, "DELETE FROM users WHERE id=?", testUserID)
	var count int
	if err = db.db.QueryRow("SELECT count(*) FROM ai_analyses").Scan(&count); err != nil || count != 0 {
		t.Fatal("account cleanup retained answer", err)
	}
}
