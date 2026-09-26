package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStandardVerifiedFieldCannotBecomeRetryableAfterExternalRevert(t *testing.T) {
	w := standardWriteFixture(t, []string{"gps", "description"})
	ctx := context.Background()
	op, err := w.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "retained-field-evidence"})
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
	w.meta["exifInfo"] = map[string]any{"latitude": 0, "longitude": 12, "description": 25}
	if err = attempt.Verify(ctx, op); err != nil {
		t.Fatal(err)
	}
	current, err := w.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || current.Status != "verifying" || len(current.Fields) != 2 || current.Fields[0].Status != "verified" {
		t.Fatal("initial partial evidence missing", err)
	}
	w.f.reopen(t)
	w.writer.drafts.results.jobs = w.f.store
	w.meta["exifInfo"] = map[string]any{"latitude": nil, "longitude": nil, "description": w.preview.Plan.Description.Before.Value}
	if err = attempt.Verify(ctx, op); err != nil {
		t.Fatal(err)
	}
	current, err = w.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || current.Status != "conflict" || current.Fields[0].Status != "conflict" {
		t.Fatal("a previously successful field regained replay authority", current.Status, err)
	}
	if _, err = w.writer.retry(ctx, testUserID, op.ID, current.Generation); err == nil {
		t.Fatal("combined payload could replay successful field")
	}
}
