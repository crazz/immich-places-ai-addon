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

func TestAIMirrorLegacyVisualProvenanceUsesTheReviewedMode(t *testing.T) {
	f := mirrorWriteFixture(t)
	ctx := context.Background()
	var mode string
	if err := f.f.db.db.QueryRow(`SELECT coalesce(json_extract(metadata,'$.mode'),'') FROM ai_analyses WHERE userID=? AND id=?`, testUserID, f.draft.AnalysisID).Scan(&mode); err != nil || mode != "" {
		t.Fatal("fixture must retain legacy omitted mode", mode, err)
	}
	d, err := f.writer.drafts.edit(ctx, testUserID, f.draft.ID, f.draft.Revision, drafts.Edit{Mirror: json.RawMessage(`{"languages":["en"],"provenance":true,"model":true}`)})
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
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "provenance"})
	if err != nil || op.Mirror == nil {
		t.Fatal("valid legacy visual provenance rejected at confirmation", err)
	}
	var export writepreview.MirrorExport
	if json.Unmarshal(op.Plan.Mirror.Value, &export) != nil || export.Provenance == nil || export.Provenance.Mode != "visual" {
		t.Fatal("reviewed mode not preserved")
	}
}
