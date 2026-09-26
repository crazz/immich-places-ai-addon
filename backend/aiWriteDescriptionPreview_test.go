package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIDescriptionPreviewPersistsExactSelectedText(t *testing.T) {
	f, image, store, value, meta := draftBaselineFixture(t)
	ctx := context.Background()
	before := strings.Repeat("Before\r\n", 4096)
	meta["exifInfo"] = map[string]any{"latitude": nil, "longitude": nil, "description": before}
	observation, err := store.observe(ctx, testUserID, value.ID, value.Revision, image.service)
	if err != nil {
		t.Fatal(err)
	}
	value, err = store.acknowledge(ctx, testUserID, value.ID, value.Revision, observation.ID, image.service)
	if err != nil {
		t.Fatal(err)
	}
	language, policy, text := "en", "replace", "Exact e\u0301\r\n scene  "
	value, err = store.edit(ctx, testUserID, value.ID, value.Revision, drafts.Edit{PrimaryLanguage: &language, DescriptionPolicy: &policy, Descriptions: map[string]drafts.DescriptionEdit{language: {Text: &text}}, Fields: json.RawMessage(`["description"]`), State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	meta["exifInfo"] = map[string]any{"latitude": 15, "longitude": 22, "description": before}
	session := &aiWritePreviewSession{drafts: store, images: image.service}
	preview, err := writepreview.Create(ctx, session, session, session, testUserID, value.ID, value.Revision, uuid.NewString(), f.store.now)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Plan.Version != "standard-preview-v2" || preview.Plan.Description == nil || preview.Plan.Description.Before.Value != before || preview.Plan.Description.Intended != text || len(preview.Plan.PolicyID) != 64 {
		t.Fatal("exact selected description missing")
	}
	f.reopen(t)
	store.results.jobs = f.store
	retained, err := session.get(ctx, testUserID, preview.Plan.ID)
	if err != nil || retained.Digest != preview.Digest || retained.Plan.Description == nil || *retained.Plan.Description != *preview.Plan.Description {
		t.Fatal("versioned text preview did not survive database reopen", err)
	}
	if _, err := session.get(ctx, "foreign", preview.Plan.ID); err == nil {
		t.Fatal("foreign owner read private text")
	}
	if _, bad := image.counts(); bad != 0 {
		t.Fatal("preview mutated Immich")
	}
}
