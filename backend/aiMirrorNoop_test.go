package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIMirrorOwnedUnchangedValueUsesVerifiedStandardAndMetadataNoops(t *testing.T) {
	f := mirrorWriteFixture(t)
	mirrorMutationServer(t, f)
	ctx := context.Background()
	d, err := f.writer.drafts.edit(ctx, testUserID, f.draft.ID, f.draft.Revision, drafts.Edit{Fields: json.RawMessage(`["gps"]`)})
	if err != nil {
		t.Fatal(err)
	}
	f.meta["exifInfo"].(map[string]any)["latitude"] = float64(0)
	f.meta["exifInfo"].(map[string]any)["longitude"] = float64(12)
	observation, err := f.writer.drafts.observe(ctx, testUserID, d.ID, d.Revision, f.image.service)
	if err != nil {
		t.Fatal(err)
	}
	d, err = f.writer.drafts.acknowledge(ctx, testUserID, d.ID, d.Revision, observation.ID, f.image.service)
	if err != nil {
		t.Fatal(err)
	}
	d, err = f.writer.drafts.edit(ctx, testUserID, d.ID, d.Revision, drafts.Edit{State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"initial-mirror", "unchanged-mirror"} {
		session := &aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service, mirrorDisclosure: writepreview.MirrorDisclosure}
		preview, err := writepreview.Create(ctx, session, session, session, testUserID, d.ID, d.Revision, uuid.NewString(), f.f.store.now)
		if err != nil {
			t.Fatal(err)
		}
		op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: key})
		if err != nil {
			t.Fatal(err)
		}
		for range 2 {
			if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
				t.Fatal(err)
			}
		}
		saved, err := f.writer.get(ctx, testUserID, op.ID, false)
		if err != nil || saved.Status != "succeeded" || !saved.Verified || !saved.Targets[0].Noop {
			t.Fatal("no-op standard prerequisite did not permit exact mirror", err)
		}
		if key == "unchanged-mirror" && (!saved.Mirror.Noop || saved.Mirror.Attempts != 0) {
			t.Fatal("unchanged owned mirror consumed mutation budget")
		}
	}
	if f.standardSends != 0 || f.mirrorSends != 1 {
		t.Fatal("no-op execution sent redundant payload", f.standardSends, f.mirrorSends)
	}
}
