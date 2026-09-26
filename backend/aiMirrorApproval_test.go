package main

import (
	"context"
	"reflect"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorConfirmationPersistsOneBlockedStepAndReplaysLocally(t *testing.T) {
	f := mirrorWriteFixture(t)
	ctx := context.Background()
	input := writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "mirror-approval"}
	op, err := f.writer.confirm(ctx, testUserID, input)
	if err != nil {
		t.Fatal(err)
	}
	if op.Mirror == nil || op.Mirror.AssetID != f.draft.AssetID || op.Mirror.Step != "metadata" || op.Mirror.Status != "blocked" || op.Mirror.Attempts != 0 || op.Mirror.Generation != 0 || len(op.Mirror.Events) != 1 || op.Mirror.Events[0].Code != "approved" || len(op.Targets) != 1 || op.Targets[0].Status != "queued" {
		t.Fatal("wrong separate metadata approval state")
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	f.writer.enabled = func() bool { return false }
	before, bad := f.image.counts()
	replay, err := f.writer.confirm(ctx, testUserID, input)
	if err != nil || !reflect.DeepEqual(op, replay) {
		t.Fatal("approval replay changed after reopen/disable", err)
	}
	after, laterBad := f.image.counts()
	if before != after || bad != 0 || laterBad != 0 {
		t.Fatal("approval replay caused I/O")
	}
	if _, err := f.writer.get(ctx, "foreign", op.ID, false); err == nil {
		t.Fatal("foreign owner read mirror operation")
	}
}
