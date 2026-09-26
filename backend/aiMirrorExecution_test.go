package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIMirrorExecutesAfterStandardAndPublishesVerifiedOwnedValue(t *testing.T) {
	f := mirrorWriteFixture(t)
	mirrorMutationServer(t, f)
	ctx := context.Background()
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "mirror-success"})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
			t.Fatal(err)
		}
	}
	op, err = f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || !op.Verified || op.Status != "succeeded" || !op.Targets[0].Verified || !op.Mirror.Verified || !op.Mirror.Settled || op.Mirror.Attempts != 1 || op.Mirror.Observed == nil {
		t.Fatal("combined success not independently verified", err)
	}
	if f.standardSends != 1 || f.mirrorSends != 1 {
		t.Fatal("steps were repeated or omitted", f.standardSends, f.mirrorSends)
	}
	record, owned, err := f.writer.drafts.mirrorRecord(ctx, testUserID, f.draft.AssetID)
	if err != nil || record != op.Plan.Mirror.RecordID || !writepreview.EqualMirrorValue(owned, op.Plan.Mirror.Value) {
		t.Fatal("verified namespace ownership was not published", err)
	}
	var guards int
	if err = f.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards`).Scan(&guards); err != nil || guards != 0 {
		t.Fatal("settled steps retained guard", err)
	}
	for range 2 {
		if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
			t.Fatal(err)
		}
	}
	if f.standardSends != 1 || f.mirrorSends != 1 {
		t.Fatal("terminal operation resent")
	}
}
