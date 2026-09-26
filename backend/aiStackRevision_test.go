package main

import (
	"context"
	"errors"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func TestAIStackDraftEditCancelsSafeTargetsAndFreezesAnyReservedMember(t *testing.T) {
	for _, reserved := range []bool{false, true} {
		f := stackWriteFixture(t)
		op := approveStackFixture(t, f, f.members)
		ctx := context.Background()
		if reserved {
			asset := op.Plan.Manifest.Targets[1].AssetID
			step, err := writeback.TargetStep(op, asset)
			if err != nil {
				t.Fatal(err)
			}
			a := &aiStackAttempt{store: f.writer, parent: op, assetID: asset}
			if _, err := a.Read(ctx, step); err != nil {
				t.Fatal(err)
			}
			if err := a.Reserve(ctx, step, false); err != nil {
				t.Fatal(err)
			}
		}
		_, err := f.writer.drafts.edit(ctx, testUserID, f.draft.ID, f.draft.Revision, drafts.Edit{State: "rejected"})
		if reserved && !errors.Is(err, drafts.ErrWriteInProgress) {
			t.Fatal("reserved sibling did not freeze draft", err)
		}
		if !reserved && err != nil {
			t.Fatal(err)
		}
		saved, err := f.writer.get(ctx, testUserID, op.ID, false)
		if err != nil {
			t.Fatal(err)
		}
		var guards, invalidated int
		if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards`).Scan(&guards); err != nil {
			t.Fatal(err)
		}
		if err := f.f.db.db.QueryRow(`SELECT invalidated FROM ai_stack_write_previews WHERE id=?`, op.Plan.ID).Scan(&invalidated); err != nil {
			t.Fatal(err)
		}
		if reserved && (saved.Status != "writing" || guards != 3 || invalidated != 0) {
			t.Fatal("blocked edit changed reserved state")
		}
		if !reserved && (saved.Status != "canceled" || guards != 0 || invalidated != 1) {
			t.Fatal("safe edit retained obsolete targets", saved.Status, guards, invalidated)
		}
	}
}
