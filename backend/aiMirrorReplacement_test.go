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

func TestAIMirrorReplacesVerifiedOwnedNamespaceWithNewlyReviewedContent(t *testing.T) {
	f := mirrorWriteFixture(t)
	mirrorMutationServer(t, f)
	ctx := context.Background()
	first, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "owned-first"})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err = f.writer.runOne(ctx, testUserID, first.ID); err != nil {
			t.Fatal(err)
		}
	}
	original := append(json.RawMessage(nil), f.namespace...)
	observation, err := f.writer.drafts.observe(ctx, testUserID, f.draft.ID, f.draft.Revision, f.image.service)
	if err != nil {
		t.Fatal(err)
	}
	d, err := f.writer.drafts.acknowledge(ctx, testUserID, f.draft.ID, f.draft.Revision, observation.ID, f.image.service)
	if err != nil {
		t.Fatal(err)
	}
	text := "Newly reviewed translation e\u0301\nPreserve exact content  "
	d, err = f.writer.drafts.edit(ctx, testUserID, d.ID, d.Revision, drafts.Edit{Descriptions: map[string]drafts.DescriptionEdit{"en": {Text: &text}}})
	if err != nil {
		t.Fatal(err)
	}
	d, err = f.writer.drafts.edit(ctx, testUserID, d.ID, d.Revision, drafts.Edit{State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	session := &aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service, mirrorDisclosure: writepreview.MirrorDisclosure}
	preview, err := writepreview.Create(ctx, session, session, session, testUserID, d.ID, d.Revision, uuid.NewString(), f.f.store.now)
	if err != nil {
		t.Fatal(err)
	}
	if !preview.Plan.Mirror.Before.Present || !writepreview.EqualMirrorValue(preview.Plan.Mirror.Before.Value, original) || preview.Plan.Mirror.RecordID != first.Plan.Mirror.RecordID {
		t.Fatal("owned before-value or stable identity changed")
	}
	second, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "owned-replacement"})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err = f.writer.runOne(ctx, testUserID, second.ID); err != nil {
			t.Fatal(err)
		}
	}
	saved, err := f.writer.get(ctx, testUserID, second.ID, false)
	if err != nil || saved.Status != "succeeded" || !saved.Mirror.Verified {
		t.Fatal("approved owned replacement failed", err)
	}
	var export writepreview.MirrorExport
	if json.Unmarshal(f.namespace, &export) != nil || export.Descriptions["en"] != text || export.RecordID != first.Plan.Mirror.RecordID {
		t.Fatal("replacement text/identity changed")
	}
	if f.standardSends != 2 || f.mirrorSends != 2 {
		t.Fatal("owned replacement duplicated mutations")
	}
	_, owned, err := f.writer.drafts.mirrorRecord(ctx, testUserID, f.draft.AssetID)
	if err != nil || !writepreview.EqualMirrorValue(owned, preview.Plan.Mirror.Value) {
		t.Fatal("new verified ownership not published", err)
	}
}
