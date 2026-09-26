package main

import (
	"context"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStackVerifiedAppendOwnershipSurvivesV3V2AndReopen(t *testing.T) {
	f := stackWriteFixture(t)
	ctx := context.Background()
	first := appendPreview(t, f.aiStandardWriteFixture, "Stack reviewed scene.")
	var err error
	f.draft, err = f.writer.drafts.get(ctx, testUserID, f.draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	op := approveStackFixture(t, f, f.members)
	stackMutationServer(t, f)
	for range 3 {
		if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
			t.Fatal(err)
		}
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	exif := f.metadata[f.draft.AssetID]["exifInfo"].(map[string]any)
	exif["description"] = op.Plan.Description.Intended + "\r\nUser suffix  "
	second := appendPreview(t, f.aiStandardWriteFixture, "Single-photo replacement.")
	if second.Plan.Description.Lineage == nil || second.Plan.Description.Lineage.ID != op.Plan.Description.Lineage.ID || strings.Count(second.Plan.Description.Intended, "[[Immich Places AI v1:") != 1 || strings.Contains(second.Plan.Description.Intended, "Stack reviewed scene.") || !strings.HasSuffix(second.Plan.Description.Intended, "\r\nUser suffix  ") {
		t.Fatal("v3 append ownership lost across v2 preview/reopen")
	}
	if first.Plan.Description.Lineage == nil {
		t.Fatal("fixture has no managed append")
	}
	v2, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: second.Plan.ID, Digest: second.Digest, Key: "v2-append-after-stack"})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.writer.runOne(ctx, testUserID, v2.ID); err != nil {
		t.Fatal(err)
	}
	third := appendPreview(t, f.aiStandardWriteFixture, "Another stack description.")
	f.draft, err = f.writer.drafts.get(ctx, testUserID, f.draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	review, err := f.writer.drafts.observeStack(ctx, testUserID, f.draft.ID, f.draft.Revision, f.members, f.image.service)
	if err != nil {
		t.Fatal(err)
	}
	preview, err := (&aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service}).createStack(ctx, testUserID, f.draft.ID, f.draft.Revision, review.ID)
	if err != nil || preview.Plan.Description.Lineage.ID != third.Plan.Description.Lineage.ID || preview.Plan.Description.Lineage.ID != op.Plan.Description.Lineage.ID || strings.Count(preview.Plan.Description.Intended, "[[Immich Places AI v1:") != 1 {
		t.Fatal("v2 append ownership unavailable to next stack plan", err)
	}
	if _, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "next-stack-append"}); err != nil {
		t.Fatal("retained owned append not confirmable", err)
	}
}
