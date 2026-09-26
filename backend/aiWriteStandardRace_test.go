package main

import (
	"context"
	"errors"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func TestAIStandardDraftEditAndReservationHaveOneAtomicWinner(t *testing.T) {
	for range 8 {
		w := standardWriteFixture(t, []string{"description"})
		ctx := context.Background()
		op, err := w.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "revision-race"})
		if err != nil {
			t.Fatal(err)
		}
		attempt := &aiWriteAttempt{store: w.writer}
		if _, err := attempt.Read(ctx, op); err != nil {
			t.Fatal(err)
		}
		start := make(chan struct{})
		edited, reserved := make(chan error, 1), make(chan error, 1)
		go func() {
			<-start
			_, err := w.writer.drafts.edit(ctx, testUserID, w.draft.ID, w.draft.Revision, drafts.Edit{State: "rejected"})
			edited <- err
		}()
		go func() { <-start; reserved <- attempt.Reserve(ctx, op, false) }()
		close(start)
		editErr, reserveErr := <-edited, <-reserved
		if (editErr == nil) == (reserveErr == nil) {
			t.Fatal("edit and reservation did not serialize", editErr, reserveErr)
		}
		current, err := w.writer.get(ctx, testUserID, op.ID, false)
		if err != nil {
			t.Fatal(err)
		}
		stored, err := w.writer.drafts.get(ctx, testUserID, w.draft.ID)
		if err != nil {
			t.Fatal(err)
		}
		var guards int
		if err := w.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards WHERE token=?`, op.ID).Scan(&guards); err != nil {
			t.Fatal(err)
		}
		if editErr == nil {
			if current.Status != "canceled" || current.Attempts != 0 || stored.Revision != w.draft.Revision+1 || guards != 0 {
				t.Fatal("edit winner left executable old approval")
			}
		} else if !errors.Is(editErr, drafts.ErrWriteInProgress) || current.Status != "writing" || current.Attempts != 1 || stored.Revision != w.draft.Revision || guards != 1 {
			t.Fatal("reservation winner lost frozen revision", editErr)
		}
	}
}
