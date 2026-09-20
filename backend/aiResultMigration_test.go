package main

import (
	"context"
	"testing"

	"github.com/pressly/goose/v3"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/review"
)

func TestAIResultUpgradeBackfillsRetainedFactsWithoutTrustingCorruptPayloads(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	lease, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok {
		t.Fatal(err)
	}
	if err = f.store.Reserve(ctx, lease); err != nil {
		t.Fatal(err)
	}
	if _, err = f.store.Complete(ctx, lease, aiJobCompletion(t)); err != nil {
		t.Fatal(err)
	}
	if err = f.store.Cancel(ctx, testUserID, job.ID); err != nil {
		t.Fatal(err)
	}
	if err = goose.DownTo(f.db.db, "migrations", 24); err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.db, `INSERT INTO ai_job_launch VALUES(?,?,?,?,?,?)`, testUserID, job.ID, selectionA, "2026-09-20T00:30:00+14:00", selectionB, "Retained album")
	selectionSQL(t, f.db, `DROP TRIGGER ai_analyses_immutable`)
	selectionSQL(t, f.db, `UPDATE ai_analyses SET payload='{"descriptions":["broken"]}'`)
	f.reopen(t)
	s := &aiResultStore{jobs: f.store}
	page, err := s.list(ctx, testUserID, review.Query{Limit: 30})
	if err != nil || len(page.Items) != 2 {
		t.Fatal("upgrade lost history", err)
	}
	for _, entry := range page.Items {
		if entry.AssetID == selectionA {
			if entry.CaptureDay == nil || *entry.CaptureDay != "2026-09-20" || entry.AlbumID == nil || *entry.AlbumID != selectionB {
				t.Fatal("retained facts lost")
			}
		} else if entry.CaptureDay != nil || entry.AlbumID != nil {
			t.Fatal("invented legacy facts")
		}
	}
}
