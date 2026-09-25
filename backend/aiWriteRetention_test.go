package main

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestAIWriteAuditSurvivesCleanupCatalogResetAndDisabledReopen(t *testing.T) {
	w := newAIWriteFixture(t)
	w.run(t)
	before := w.status(t)
	ctx := context.Background()
	analysis, err := w.f.store.ReadAnalysis(ctx, testUserID, w.draft.AnalysisID)
	if err != nil {
		t.Fatal(err)
	}
	if err = w.f.store.Cancel(ctx, testUserID, analysis.JobID); err != nil {
		t.Fatal(err)
	}
	w.f.now = w.f.now.Add(48 * time.Hour)
	selectionSQL(t, w.f.db, "DELETE FROM assets WHERE userID=?", testUserID)
	purged, err := w.f.store.PurgeBefore(ctx, testUserID, w.f.now.Add(-time.Hour), 100)
	if err != nil || purged != 0 {
		t.Fatal(purged, err)
	}
	selectionSQL(t, w.f.db, "DELETE FROM ai_write_previews WHERE protected=0 AND expiresAt<?", w.f.now.UnixNano())
	w.f.reopen(t)
	w.writer.drafts.results.jobs = w.f.store
	w.enabled = false
	after := w.status(t)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("private audit changed during cleanup")
	}
	if _, err = w.writer.drafts.get(ctx, testUserID, w.draft.ID); err != nil {
		t.Fatal("referenced draft lost", err)
	}
	if _, err = w.f.store.ReadAnalysis(ctx, testUserID, w.draft.AnalysisID); err != nil {
		t.Fatal("referenced result lost", err)
	}
	w.run(t)
	if sends, _ := w.counts(); sends != 1 {
		t.Fatal("cleanup resent", sends)
	}
}
