package main

import (
	"context"
	"encoding/json"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStandardConfirmationAndReloadHaveIdenticalPendingFieldProjection(t *testing.T) {
	w := standardWriteFixture(t, []string{"description"})
	ctx := context.Background()
	op, err := w.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "stable-api"})
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := w.writer.get(ctx, testUserID, op.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	created, e1 := json.Marshal(op)
	retained, e2 := json.Marshal(loaded)
	if e1 != nil || e2 != nil || string(created) != string(retained) {
		t.Fatal("confirmation and reload disagree on operation projection")
	}
	if len(op.Fields) != 1 || op.Fields[0].Field != "description" || op.Fields[0].Status != "pending" {
		t.Fatal("selected pending field missing")
	}
}
