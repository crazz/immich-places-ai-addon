package main

import (
	"context"
	"testing"
	"time"
)

func TestAITranslationPrivateOperationsAndRetentionAcrossOwners(t *testing.T) {
	f, s, req := translationFixture(t)
	ctx := context.Background()
	own, err := s.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	draft, err := s.drafts.get(ctx, testUserID, req.DraftID)
	if err != nil {
		t.Fatal(err)
	}
	analysis, err := f.store.ReadAnalysis(ctx, testUserID, draft.AnalysisID)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.store.Cancel(ctx, testUserID, analysis.JobID); err != nil {
		t.Fatal(err)
	}
	if err = f.db.createUser(ctx, "other-owner", "translation-other@example.test", "hash"); err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.createAIProvider(ctx, "other-owner", "profile", providerInput()); err != nil {
		t.Fatal(err)
	}
	f.input.Owner = "other-owner"
	job := completeAIJob(t, f, "other-translation")
	var analysisID string
	if err = f.db.db.QueryRow("SELECT id FROM ai_analyses WHERE userID=? AND jobID=?", "other-owner", job.ID).Scan(&analysisID); err != nil {
		t.Fatal(err)
	}
	otherDraft, err := s.drafts.accept(ctx, "other-owner", analysisID, nil)
	if err != nil {
		t.Fatal(err)
	}
	otherReq := req
	otherReq.DraftID = otherDraft.ID
	otherReq.Revision = otherDraft.Revision
	otherReq.FactsRevision = otherDraft.FactsRevision
	other, err := s.submit(ctx, "other-owner", otherReq)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.get(ctx, "other-owner", own.ID); err == nil {
		t.Fatal("foreign read")
	}
	if _, err = s.list(ctx, "other-owner", req.DraftID, ""); err == nil {
		t.Fatal("foreign history")
	}
	if err = s.cancel(ctx, "other-owner", own.ID); err == nil {
		t.Fatal("foreign cancellation")
	}
	if _, err = s.adopt(ctx, "other-owner", own.ID, req.Revision, []string{"en"}); err == nil {
		t.Fatal("foreign adoption")
	}
	selectionSQL(t, f.db, "UPDATE ai_translation_items SET state='failed'")
	otherReq.Key = "foreign-parent"
	otherReq.ParentID = own.ID
	if _, err = s.submit(ctx, "other-owner", otherReq); err == nil {
		t.Fatal("foreign parent retry")
	}
	f.now = f.now.Add(2 * time.Hour)
	selectionSQL(t, f.db, "DELETE FROM assets")
	for _, owner := range []string{testUserID, "other-owner"} {
		if n, e := f.store.PurgeBefore(ctx, owner, f.now.Add(-time.Hour), 100); e != nil || n != 0 {
			t.Fatal("draft-linked history purged", n, e)
		}
	}
	if _, err = s.get(ctx, testUserID, own.ID); err != nil {
		t.Fatal("catalog reset lost retained translation", err)
	}
	selectionSQL(t, f.db, "DELETE FROM users WHERE ID=?", testUserID)
	if _, err = s.get(ctx, testUserID, own.ID); err == nil {
		t.Fatal("deleted translation survived")
	}
	if _, err = s.get(ctx, "other-owner", other.ID); err != nil {
		t.Fatal("other history removed", err)
	}
}
