package main

import (
	"context"
	"encoding/json"
	"testing"

	"immich-places-backend/internal/ai/drafts"
)

func TestAIMirrorSelectionCannotSilentlyDowngradeToStandardOnlyPreview(t *testing.T) {
	f := standardWriteFixture(t, []string{"gps"})
	ctx := context.Background()
	d, err := f.writer.drafts.edit(ctx, testUserID, f.draft.ID, f.draft.Revision, drafts.Edit{Mirror: json.RawMessage(`{"provenance":true}`)})
	if err != nil {
		t.Fatal(err)
	}
	session := &aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service}
	if _, err = session.ReadDraft(ctx, testUserID, d.ID); err == nil {
		t.Fatal("unsupported mirror selection silently became standard-only snapshot")
	}
	d, err = f.writer.drafts.edit(ctx, testUserID, d.ID, d.Revision, drafts.Edit{Mirror: json.RawMessage(`null`)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = session.ReadDraft(ctx, testUserID, d.ID); err != nil {
		t.Fatal("unsupported optional mirror blocked standard-only selection", err)
	}
}
