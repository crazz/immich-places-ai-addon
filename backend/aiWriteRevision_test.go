package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func TestAIWriteDraftEditCancelsQueuedButProtectsReservedRevision(t *testing.T) {
	for _, reserved := range []bool{false, true} {
		t.Run(map[bool]string{false: "queued", true: "reserved"}[reserved], func(t *testing.T) {
			w := newAIWriteFixture(t)
			if reserved {
				a := &aiWriteAttempt{store: w.writer}
				if _, err := a.Read(context.Background(), w.op); err != nil {
					t.Fatal(err)
				}
				if err := a.Reserve(context.Background(), w.op, false); err != nil {
					t.Fatal(err)
				}
			}
			value, err := w.writer.drafts.edit(context.Background(), testUserID, w.draft.ID, w.draft.Revision, drafts.Edit{Camera: json.RawMessage(`{"latitude":3,"longitude":4}`)})
			if reserved {
				if !errors.Is(err, drafts.ErrWriteInProgress) {
					t.Fatal("active revision edited", value, err)
				}
			} else {
				if err != nil || value.Revision != w.draft.Revision+1 {
					t.Fatal(value, err)
				}
				if op := w.status(t); op.Status != "canceled" {
					t.Fatal("approval survived edit", op)
				}
				if _, err = w.writer.confirm(context.Background(), testUserID, writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "new-key"}); err == nil {
					t.Fatal("old approval reused")
				}
			}
		})
	}
}

func TestAIWriteRetryableApprovalIsCanceledByDraftRevision(t *testing.T) {
	w := newAIWriteFixture(t)
	selectionSQL(t, w.f.db, "UPDATE ai_write_operations SET status='retryable' WHERE id=?", w.op.ID)
	selectionSQL(t, w.f.db, "UPDATE ai_write_targets SET attempts=1,completionKnown=1,senderActive=0 WHERE operationID=?", w.op.ID)
	value, err := w.writer.drafts.edit(context.Background(), testUserID, w.draft.ID, w.draft.Revision, drafts.Edit{Camera: json.RawMessage(`{"latitude":3,"longitude":4}`)})
	if err != nil || value.Revision != w.draft.Revision+1 {
		t.Fatal(value, err)
	}
	if op := w.status(t); op.Status != "canceled" || op.Attempts != 1 {
		t.Fatal(op)
	}
	if err := w.writer.runOne(context.Background(), testUserID, w.op.ID); err != nil {
		t.Fatal(err)
	}
	if sends, _ := w.counts(); sends != 0 {
		t.Fatal("canceled approval sent", sends)
	}
}
