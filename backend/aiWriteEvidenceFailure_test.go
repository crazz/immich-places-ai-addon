package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStandardCatalogFailureRetainsVerifiedEvidenceBeforeExternalRevert(t *testing.T) {
	w := standardWriteFixture(t, []string{"gps", "description"})
	ctx := context.Background()
	op, err := w.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "refresh-failure-evidence"})
	if err != nil {
		t.Fatal(err)
	}
	attempt := &aiWriteAttempt{store: w.writer}
	if _, err = attempt.Read(ctx, op); err != nil {
		t.Fatal(err)
	}
	if err = attempt.Reserve(ctx, op, false); err != nil {
		t.Fatal(err)
	}
	if err = attempt.Sent(ctx, op, writeback.Completion{Known: true}); err != nil {
		t.Fatal(err)
	}
	w.meta["exifInfo"] = map[string]any{"latitude": 0, "longitude": 12, "description": op.Plan.Description.Intended}
	selectionSQL(t, w.f.db, `CREATE TRIGGER fail_standard_refresh BEFORE UPDATE OF latitude ON assets BEGIN SELECT RAISE(ABORT,'synthetic refresh failure'); END`)
	if err = attempt.Verify(ctx, op); err == nil {
		t.Fatal("expected local refresh failure")
	}
	w.f.reopen(t)
	w.writer.drafts.results.jobs = w.f.store
	current, err := w.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || current.Status != "verifying" || current.Refreshed || len(current.Fields) != 2 || !current.Fields[0].WasVerified || !current.Fields[1].WasVerified {
		t.Fatal("local refresh failure discarded verified evidence", current.Fields, err)
	}
	selectionSQL(t, w.f.db, "DROP TRIGGER fail_standard_refresh")
	w.meta["exifInfo"] = map[string]any{"latitude": nil, "longitude": nil, "description": op.Plan.Description.Before.Value}
	if err = attempt.Verify(ctx, op); err != nil {
		t.Fatal(err)
	}
	current, err = w.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || current.Status != "conflict" || current.Attempts != 1 {
		t.Fatal("external revert restored replay authority", current.Status, err)
	}
	if _, err = w.writer.retry(ctx, testUserID, op.ID, current.Generation); err == nil {
		t.Fatal("previously verified combined request could replay")
	}
}
