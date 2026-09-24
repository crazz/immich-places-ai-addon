package main

import (
	"context"
	"testing"
	"time"
)

func TestAIDraftProtectsHistoryFromCleanupAndCatalogReset(t *testing.T) {
	f, id := draftFixture(t)
	value := acceptedDraft(t, f, id)
	ctx := context.Background()
	analysis, err := f.store.ReadAnalysis(ctx, testUserID, id)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.store.Cancel(ctx, testUserID, analysis.JobID); err != nil {
		t.Fatal(err)
	}
	f.now = f.now.Add(2 * time.Hour)
	selectionSQL(t, f.db, "DELETE FROM assets WHERE userID=?", testUserID)
	f.store.enabled = false
	deleted, err := f.store.PurgeBefore(ctx, testUserID, f.now.Add(-time.Hour), 100)
	if err != nil || deleted != 0 {
		t.Fatalf("retained draft cleanup: %d %v", deleted, err)
	}
	rec := aiRequest(draftHandler(f), "GET", "/ai/drafts/"+value.ID, "", "", true)
	if rec.Code != 200 {
		t.Fatal("draft lost", rec.Code)
	}
	detail, err := (&aiResultStore{jobs: f.store}).detail(ctx, testUserID, id, "", "")
	if err != nil || detail.Proposal == nil {
		t.Fatal("referenced analysis lost", err)
	}
}

func TestAIDraftDeniesForeignAndObsoleteScope(t *testing.T) {
	f, id := draftFixture(t)
	value := acceptedDraft(t, f, id)
	ctx := context.Background()
	s := &aiDraftStore{results: &aiResultStore{jobs: f.store}}
	if _, err := s.get(ctx, "foreign", value.ID); err == nil {
		t.Fatal("foreign read")
	}
	if _, err := s.accept(ctx, "foreign", id, nil); err == nil {
		t.Fatal("foreign accept")
	}
	rec := aiRequest(draftHandler(f), "GET", "/ai/drafts/"+value.ID, "", "", false)
	if rec.Code != 401 {
		t.Fatal("anonymous read", rec.Code)
	}
	selectionSQL(t, f.db, "UPDATE ai_installation_identity SET id=?", selectionB)
	if _, err := s.get(ctx, testUserID, value.ID); err == nil {
		t.Fatal("obsolete installation")
	}
	s.results.jobs = newAIJobStore(f.db, selectionB, false, f.store.now)
	if _, err := s.get(ctx, testUserID, value.ID); err == nil {
		t.Fatal("old draft reauthorized")
	}
}

func TestAIDraftAccountDeletionCascadesWithoutLateRecreation(t *testing.T) {
	f, id := draftFixture(t)
	value := acceptedDraft(t, f, id)
	ctx := context.Background()
	original, err := f.store.ReadAnalysis(ctx, testUserID, id)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.store.Cancel(ctx, testUserID, original.JobID); err != nil {
		t.Fatal(err)
	}
	if err = f.db.createUser(ctx, "other-owner", "draft-other@example.com", "hash"); err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.createAIProvider(ctx, "other-owner", "profile", providerInput()); err != nil {
		t.Fatal(err)
	}
	f.input.Owner = "other-owner"
	otherJob := completeAIJob(t, f, "other-draft")
	var otherID string
	if err = f.db.db.QueryRow("SELECT id FROM ai_analyses WHERE userID=? AND jobID=?", "other-owner", otherJob.ID).Scan(&otherID); err != nil {
		t.Fatal(err)
	}
	store := &aiDraftStore{results: &aiResultStore{jobs: f.store}}
	other, err := store.accept(ctx, "other-owner", otherID, nil)
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.db, "DELETE FROM users WHERE ID=?", testUserID)
	if _, err = store.accept(ctx, testUserID, id, nil); err == nil {
		t.Fatal("late accept recreated deleted account")
	}
	if _, err = store.get(ctx, testUserID, value.ID); err == nil {
		t.Fatal("deleted draft survived")
	}
	if _, err = store.get(ctx, "other-owner", other.ID); err != nil {
		t.Fatal("other draft lost", err)
	}
	var count int
	if err = f.db.db.QueryRow("SELECT count(*) FROM ai_draft_revisions WHERE userID=?", testUserID).Scan(&count); err != nil || count != 0 {
		t.Fatal("private revisions retained", count, err)
	}
}

func TestAIDraftAcceptanceSerializesWithPurge(t *testing.T) {
	for range 8 {
		f, id := draftFixture(t)
		ctx := context.Background()
		original, err := f.store.ReadAnalysis(ctx, testUserID, id)
		if err != nil {
			t.Fatal(err)
		}
		if err = f.store.Cancel(ctx, testUserID, original.JobID); err != nil {
			t.Fatal(err)
		}
		f.now = f.now.Add(2 * time.Hour)
		store := &aiDraftStore{results: &aiResultStore{jobs: f.store}}
		start := make(chan struct{})
		accepted := make(chan error, 1)
		purged := make(chan error, 1)
		go func() { <-start; _, err := store.accept(ctx, testUserID, id, nil); accepted <- err }()
		go func() {
			<-start
			_, err := f.store.PurgeBefore(ctx, testUserID, f.now.Add(-time.Hour), 100)
			purged <- err
		}()
		close(start)
		acceptErr, purgeErr := <-accepted, <-purged
		if purgeErr != nil {
			t.Fatal("cleanup conflict", purgeErr)
		}
		var drafts, analyses, violations int
		if err = f.db.db.QueryRow("SELECT count(*) FROM ai_drafts WHERE analysisID=?", id).Scan(&drafts); err != nil {
			t.Fatal(err)
		}
		if err = f.db.db.QueryRow("SELECT count(*) FROM ai_analyses WHERE id=?", id).Scan(&analyses); err != nil {
			t.Fatal(err)
		}
		if err = f.db.db.QueryRow("SELECT count(*) FROM pragma_foreign_key_check").Scan(&violations); err != nil {
			t.Fatal(err)
		}
		if violations != 0 || drafts != analyses || (acceptErr == nil && drafts != 1) {
			t.Fatal("orphan or lost accepted draft", drafts, analyses, acceptErr)
		}
	}
}

func TestAIDraftReanalysisKeepsEachDecisionDistinct(t *testing.T) {
	f, id := draftFixture(t)
	original := acceptedDraft(t, f, id)
	ctx := context.Background()
	analysis, err := f.store.ReadAnalysis(ctx, testUserID, id)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.store.Cancel(ctx, testUserID, analysis.JobID); err != nil {
		t.Fatal(err)
	}
	newerJob := completeAIJob(t, f, "newer-draft-run")
	var newerID string
	if err = f.db.db.QueryRow("SELECT id FROM ai_analyses WHERE userID=? AND jobID=?", testUserID, newerJob.ID).Scan(&newerID); err != nil {
		t.Fatal(err)
	}
	newer := acceptedDraft(t, f, newerID)
	if newer.ID == original.ID || newer.AssetID != original.AssetID {
		t.Fatal("run identity conflated")
	}
	if rec := draftPatch(f, original.ID, `"1"`, `{"state":"rejected"}`); rec.Code != 200 {
		t.Fatal(rec.Code)
	}
	if rec := draftPatch(f, original.ID, `"2"`, `{"state":"draft"}`); rec.Code != 200 {
		t.Fatal(rec.Code)
	}
	if got := acceptedDraft(t, f, newerID); got.Revision != 1 || got.State != "draft" {
		t.Fatal("newer draft overwritten", got)
	}
}
