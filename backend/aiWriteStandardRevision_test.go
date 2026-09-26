package main

import (
	"context"
	"errors"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func TestAIStandardDraftEditCancelsQueuedAndFreezesReserved(t *testing.T) {
	for _, reserved := range []bool{false, true} {
		w := standardWriteFixture(t, []string{"description"})
		ctx := context.Background()
		op, err := w.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "standard-edit-guard"})
		if err != nil {
			t.Fatal(err)
		}
		if reserved {
			attempt := &aiWriteAttempt{store: w.writer}
			if _, err = attempt.Read(ctx, op); err != nil {
				t.Fatal(err)
			}
			if err = attempt.Reserve(ctx, op, false); err != nil {
				t.Fatal(err)
			}
		}
		_, err = w.writer.drafts.edit(ctx, testUserID, w.draft.ID, w.draft.Revision, drafts.Edit{State: "rejected"})
		if reserved && !errors.Is(err, drafts.ErrWriteInProgress) {
			t.Fatal("possibly acting v2 write did not freeze revision", err)
		}
		if !reserved && err != nil {
			t.Fatal(err)
		}
		saved, err := w.writer.get(ctx, testUserID, op.ID, false)
		if err != nil {
			t.Fatal(err)
		}
		if !reserved && saved.Status != "canceled" {
			t.Fatal("queued v2 approval survived draft edit", saved.Status)
		}
		var guards int
		if err = w.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards WHERE token=?`, op.ID).Scan(&guards); err != nil {
			t.Fatal(err)
		}
		if (guards == 1) != reserved {
			t.Fatal("incorrect cross-version guard retention", guards)
		}
	}
}
