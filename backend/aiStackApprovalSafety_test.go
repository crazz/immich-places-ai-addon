package main

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/writeback"
)

func TestAIStackOverlapRollsBackAllNewApprovalAndGuards(t *testing.T) {
	f := stackWriteFixture(t)
	ctx := context.Background()
	review, err := f.writer.drafts.observeStack(ctx, testUserID, f.draft.ID, f.draft.Revision, f.members, f.image.service)
	if err != nil {
		t.Fatal(err)
	}
	preview, err := (&aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service}).createStack(ctx, testUserID, f.draft.ID, f.draft.Revision, review.ID)
	if err != nil {
		t.Fatal(err)
	}
	last := preview.Plan.Manifest.Targets[len(preview.Plan.Manifest.Targets)-1].AssetID
	token := uuid.NewString()
	if _, err := f.f.db.db.Exec(`INSERT INTO ai_write_target_guards(installationID,assetID,token) VALUES(?,?,?)`, preview.Plan.Installation, last, token); err != nil {
		t.Fatal(err)
	}
	_, err = f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "overlap"})
	if err == nil || err.Error() != "TARGET_BUSY" {
		t.Fatal("expected private-safe overlap rejection", err)
	}
	for _, table := range []string{"ai_stack_write_operations", "ai_stack_write_targets", "ai_stack_write_events"} {
		var count int
		if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count); err != nil || count != 0 {
			t.Fatal("partial approval survived", table, count, err)
		}
	}
	var guards, protected int
	var retained string
	if err := f.f.db.db.QueryRow(`SELECT count(*),min(token) FROM ai_write_target_guards`).Scan(&guards, &retained); err != nil || guards != 1 || retained != token {
		t.Fatal("new guards leaked or foreign guard altered", err)
	}
	if err := f.f.db.db.QueryRow(`SELECT protected FROM ai_stack_write_previews WHERE id=?`, preview.Plan.ID).Scan(&protected); err != nil || protected != 0 {
		t.Fatal("failed approval consumed preview", err)
	}
	_, writes := f.image.counts()
	if writes != 0 {
		t.Fatal("approval mutated Immich")
	}
}
