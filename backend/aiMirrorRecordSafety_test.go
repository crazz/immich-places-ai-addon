package main

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorRecordRejectsMissingAndUnverifiedStepLineage(t *testing.T) {
	f := mirrorWriteFixture(t)
	ctx := context.Background()
	update := `UPDATE ai_mirror_records SET lastOperationID=?,value=?,verifiedAt=1 WHERE userID=? AND installationID=? AND assetID=?`
	if _, err := f.f.db.db.Exec(update, uuid.NewString(), string(f.preview.Plan.Mirror.Value), testUserID, f.f.store.binding, f.draft.AssetID); err == nil {
		t.Fatal("missing lineage admitted")
	}
	_, owned, err := f.writer.drafts.mirrorRecord(ctx, testUserID, f.draft.AssetID)
	if err != nil || len(owned) != 0 {
		t.Fatal("failed update fabricated ownership", err)
	}
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "unverified-record"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.f.db.db.Exec(update, op.ID, string(op.Plan.Mirror.Value), testUserID, f.f.store.binding, f.draft.AssetID); err != nil {
		t.Fatal(err)
	}
	if _, _, err = f.writer.drafts.mirrorRecord(ctx, testUserID, f.draft.AssetID); err == nil {
		t.Fatal("unverified metadata step granted ownership")
	}
}
