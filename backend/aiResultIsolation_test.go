package main

import (
	"context"
	"immich-places-backend/internal/ai/review"
	"testing"
)

func TestAIResultOwnerAndInstallationIsolationIncludesEveryRead(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.store.Cancel(ctx, testUserID, job.ID); err != nil {
		t.Fatal(err)
	}
	s := &aiResultStore{jobs: f.store}
	page, err := s.list(ctx, "foreign", review.Query{Limit: 30})
	if err != nil || len(page.Items) != 0 {
		t.Fatal("foreign history disclosed")
	}
	if _, err = s.detail(ctx, "foreign", "", job.ID, job.Items[0].ID); err == nil {
		t.Fatal("foreign detail disclosed")
	}
	selectionSQL(t, f.db, `UPDATE ai_installation_identity SET id=?`, selectionB)
	if _, err = s.list(ctx, testUserID, review.Query{Limit: 30}); err == nil {
		t.Fatal("stale reader installation accepted")
	}
	s.jobs = newAIJobStore(f.db, selectionB, false, f.store.now)
	page, err = s.list(ctx, testUserID, review.Query{Limit: 30})
	if err != nil || len(page.Items) != 0 {
		t.Fatal("old installation history disclosed")
	}
	if _, err = s.detail(ctx, testUserID, "", job.ID, job.Items[0].ID); err == nil {
		t.Fatal("old installation detail disclosed")
	}
}
