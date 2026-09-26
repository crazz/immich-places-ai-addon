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

type aiStandardWriteFixture struct {
	f       *aiJobFixture
	image   *aiImageFixture
	writer  *aiWriteStore
	draft   drafts.Draft
	preview writepreview.Preview
	meta    map[string]any
}

func standardWriteFixture(t *testing.T, fields []string) *aiStandardWriteFixture {
	t.Helper()
	f, image, store, value, meta := draftBaselineFixture(t)
	ctx := context.Background()
	meta["exifInfo"] = map[string]any{"latitude": nil, "longitude": nil, "description": "User text\r\n"}
	writer := &aiWriteStore{drafts: store, images: image.service, sync: newSyncService(f.db, nil, nil), enabled: func() bool { return true }, profile: "immich-v3.2.2"}
	writer.capabilities = writeback.CapabilityPolicy{Version: 1, Installation: f.store.binding, Profile: writer.profile, Evidence: "synthetic-test", Capabilities: []string{"description"}}
	store.results.writer = writer
	observation, err := store.observe(ctx, testUserID, value.ID, value.Revision, image.service)
	if err != nil {
		t.Fatal(err)
	}
	value, err = store.acknowledge(ctx, testUserID, value.ID, value.Revision, observation.ID, image.service)
	if err != nil {
		t.Fatal(err)
	}
	language, policy, text := "en", "replace", "Exact e\u0301\r\n scene  "
	selection, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	value, err = store.edit(ctx, testUserID, value.ID, value.Revision, drafts.Edit{PrimaryLanguage: &language, DescriptionPolicy: &policy, Descriptions: map[string]drafts.DescriptionEdit{language: {Text: &text}}, Camera: json.RawMessage(`{"latitude":0,"longitude":12}`), Fields: selection, State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	session := &aiWritePreviewSession{drafts: store, images: image.service}
	preview, err := writepreview.Create(ctx, session, session, session, testUserID, value.ID, value.Revision, uuid.NewString(), f.store.now)
	if err != nil {
		t.Fatal(err)
	}
	return &aiStandardWriteFixture{f: f, image: image, writer: writer, draft: value, preview: preview, meta: meta}
}
