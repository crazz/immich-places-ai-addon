package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func TestAIWriteUnchangedPlanVerifiesWithoutMutation(t *testing.T) {
	f, image, store, draft, meta := writePreviewFixture(t)
	ctx := context.Background()
	meta["exifInfo"] = map[string]float64{"latitude": 0, "longitude": 12}
	observation, err := store.observe(ctx, testUserID, draft.ID, draft.Revision, image.service)
	if err != nil {
		t.Fatal(err)
	}
	draft, err = store.acknowledge(ctx, testUserID, draft.ID, draft.Revision, observation.ID, image.service)
	if err != nil {
		t.Fatal(err)
	}
	draft, err = store.edit(ctx, testUserID, draft.ID, draft.Revision, drafts.Edit{State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	preview := savedWritePreview(t, f, image, draft)
	writer := &aiWriteStore{drafts: store, images: image.service, sync: newSyncService(f.db, nil, nil), enabled: func() bool { return true }, profile: "immich-v3.2.2"}
	op, err := writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "noop"})
	if err != nil {
		t.Fatal(err)
	}
	_, badBefore := image.counts()
	if err = writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	op, err = writer.get(ctx, testUserID, op.ID, false)
	if err != nil || op.Status != "succeeded" || !op.Noop || !op.Verified || !op.Refreshed || op.Attempts != 0 {
		t.Fatal(op, err)
	}
	if _, bad := image.counts(); bad != badBefore {
		t.Fatal("mutation sent for no-op")
	}
}
