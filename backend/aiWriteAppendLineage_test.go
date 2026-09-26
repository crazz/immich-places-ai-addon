package main

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func appendPreview(t *testing.T, w *aiStandardWriteFixture, text string) writepreview.Preview {
	t.Helper()
	ctx := context.Background()
	store := w.writer.drafts
	value, err := store.get(ctx, testUserID, w.draft.ID)
	if err != nil {
		t.Fatal(err)
	}
	observation, err := store.observe(ctx, testUserID, value.ID, value.Revision, w.image.service)
	if err != nil {
		t.Fatal(err)
	}
	value, err = store.acknowledge(ctx, testUserID, value.ID, value.Revision, observation.ID, w.image.service)
	if err != nil {
		t.Fatal(err)
	}
	policy := "managed_append"
	value, err = store.edit(ctx, testUserID, value.ID, value.Revision, drafts.Edit{DescriptionPolicy: &policy, Descriptions: map[string]drafts.DescriptionEdit{"en": {Text: &text}}, State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	session := &aiWritePreviewSession{drafts: store, images: w.image.service}
	preview, err := writepreview.Create(ctx, session, session, session, testUserID, value.ID, value.Revision, uuid.NewString(), w.f.store.now)
	if err != nil {
		t.Fatal(err)
	}
	return preview
}

func TestAIVerifiedAppendOwnsStableLineageAcrossReopenAndNewDescriptions(t *testing.T) {
	w := standardWriteFixture(t, []string{"description"})
	ctx := context.Background()
	first := appendPreview(t, w, "First reviewed scene.")
	op, err := w.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: first.Plan.ID, Digest: first.Digest, Key: "first-append"})
	if err != nil {
		t.Fatal(err)
	}
	attempt := &aiWriteAttempt{store: w.writer}
	if _, err = attempt.Read(ctx, op); err != nil {
		t.Fatal(err)
	}
	if err = attempt.Reserve(ctx, op, false); err != nil {
		t.Fatal(err)
	}
	if err = attempt.Sent(ctx, op, writeback.Completion{Known: true}); err != nil {
		t.Fatal(err)
	}
	w.meta["exifInfo"] = map[string]any{"description": first.Plan.Description.Intended}
	if err = attempt.Verify(ctx, op); err != nil {
		t.Fatal(err)
	}
	w.f.reopen(t)
	w.writer.drafts.results.jobs = w.f.store
	w.meta["exifInfo"] = map[string]any{"description": first.Plan.Description.Intended + "\r\nUser suffix  "}
	second := appendPreview(t, w, "A newly reviewed scene.")
	if second.Plan.Description.Lineage.ID != first.Plan.Description.Lineage.ID || strings.Count(second.Plan.Description.Intended, "[[Immich Places AI v1:") != 1 || !strings.HasPrefix(second.Plan.Description.Intended, "User text\r\n\n\n") || !strings.HasSuffix(second.Plan.Description.Intended, "\r\nUser suffix  ") || strings.Contains(second.Plan.Description.Intended, "First reviewed scene.") {
		t.Fatal("owned block duplicated or surrounding user text changed")
	}
	if _, err = w.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: second.Plan.ID, Digest: second.Digest, Key: "second-append"}); err != nil {
		t.Fatal("repeat append not confirmable", err)
	}
}
