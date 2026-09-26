package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorOwnershipRequiresExactApprovedAndObservedValue(t *testing.T) {
	for _, tamper := range []string{"record", "observation"} {
		t.Run(tamper, func(t *testing.T) {
			f := mirrorWriteFixture(t)
			mirrorMutationServer(t, f)
			ctx := context.Background()
			op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "lineage"})
			if err != nil {
				t.Fatal(err)
			}
			for range 2 {
				if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
					t.Fatal(err)
				}
			}
			if tamper == "record" {
				selectionSQL(t, f.f.db, `UPDATE ai_mirror_records SET value='{}' WHERE userID=?`, testUserID)
			} else {
				selectionSQL(t, f.f.db, `UPDATE ai_mirror_write_steps SET observed='{"present":true,"value":{}}' WHERE userID=?`, testUserID)
			}
			if _, _, err = f.writer.drafts.mirrorRecord(ctx, testUserID, f.draft.AssetID); err == nil {
				t.Fatal("unproven local value granted namespace ownership")
			}
		})
	}
}
