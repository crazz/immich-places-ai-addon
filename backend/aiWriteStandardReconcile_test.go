package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStandardReconciliationResetsReadBudgetWithoutMutationAuthority(t *testing.T) {
	w := standardWriteFixture(t, []string{"description"})
	ctx := context.Background()
	op, err := w.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "standard-reconcile"})
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, w.f.db, `UPDATE ai_standard_write_operations SET status='verifying' WHERE id=?`, op.ID)
	selectionSQL(t, w.f.db, `UPDATE ai_standard_write_targets SET reads=3,attempts=1,completionKnown=0 WHERE operationID=?`, op.ID)
	reads, bad := w.image.counts()
	if _, err = w.writer.reconcile(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	var remaining, attempts int
	if err = w.f.db.db.QueryRow(`SELECT reads,attempts FROM ai_standard_write_targets WHERE operationID=?`, op.ID).Scan(&remaining, &attempts); err != nil || remaining != 0 || attempts != 1 {
		t.Fatal("v2 reconcile lost bounded read-only state", remaining, attempts, err)
	}
	if next, invalid := w.image.counts(); next != reads || invalid != bad {
		t.Fatal("reconcile made upstream request")
	}
}
