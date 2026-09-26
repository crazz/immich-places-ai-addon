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

func TestAIStandardUnknownSenderExcludesAnotherOwnersRemainingV1Plan(t *testing.T) {
	w := standardWriteFixture(t, []string{"gps", "description"})
	ctx := context.Background()
	op, err := w.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "unknown-combined"})
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
	w.meta["exifInfo"] = map[string]any{"latitude": nil, "longitude": nil, "description": op.Plan.Description.Intended}
	if err = attempt.Verify(ctx, op); err != nil {
		t.Fatal(err)
	}
	current, err := w.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || current.Status != "verifying" || current.Fields[1].Status != "verified" {
		t.Fatal("unknown mixed outcome settled", err)
	}
	other := "remaining-field-owner"
	analysis, err := w.f.store.ReadAnalysis(ctx, testUserID, w.draft.AnalysisID)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.f.store.Cancel(ctx, testUserID, analysis.JobID); err != nil {
		t.Fatal(err)
	}
	if err = w.f.db.createUser(ctx, other, "remaining@example.com", "hash"); err != nil {
		t.Fatal(err)
	}
	key := "private-immich-image-key"
	if err = w.f.db.updateImmichAPIKey(ctx, other, &key); err != nil {
		t.Fatal(err)
	}
	if err = w.f.db.upsertAssets(ctx, other, []AssetRow{{ImmichID: op.Plan.TargetID, Type: "IMAGE", OriginalFileName: "synthetic.jpg", FileCreatedAt: "2026-09-20"}}); err != nil {
		t.Fatal(err)
	}
	if _, err = w.f.db.createAIProvider(ctx, other, "profile", providerInput()); err != nil {
		t.Fatal(err)
	}
	w.f.input.Owner, w.f.input.AssetIDs, w.f.input.MaxCalls = other, []string{op.Plan.TargetID}, 3
	job := completeAIJob(t, w.f, "remaining-gps")
	var analysisID string
	if err = w.f.db.db.QueryRow("SELECT id FROM ai_analyses WHERE userID=? AND jobID=?", other, job.ID).Scan(&analysisID); err != nil {
		t.Fatal(err)
	}
	store := w.writer.drafts
	draft, err := store.accept(ctx, other, analysisID, nil)
	if err != nil {
		t.Fatal(err)
	}
	observation, err := store.observe(ctx, other, draft.ID, draft.Revision, w.image.service)
	if err != nil {
		t.Fatal(err)
	}
	draft, err = store.acknowledge(ctx, other, draft.ID, draft.Revision, observation.ID, w.image.service)
	if err != nil {
		t.Fatal(err)
	}
	draft, err = store.edit(ctx, other, draft.ID, draft.Revision, drafts.Edit{Camera: json.RawMessage(`{"latitude":0,"longitude":12}`), Fields: json.RawMessage(`["gps"]`), State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	session := &aiWritePreviewSession{drafts: store, images: w.image.service}
	preview, err := writepreview.Create(ctx, session, session, session, other, draft.ID, draft.Revision, uuid.NewString(), w.f.store.now)
	if err != nil || preview.Plan.Version != "gps-preview-v1" {
		t.Fatal("remaining GPS preview", err)
	}
	w.f.reopen(t)
	store.results.jobs = w.f.store
	input := writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "remaining-approved"}
	reads, bad := w.image.counts()
	if _, err := w.writer.confirm(ctx, other, input); err == nil || err.Error() != "TARGET_BUSY" {
		t.Fatal("unknown v2 sender failed to exclude v1 owner", err)
	}
	if _, err := w.writer.get(ctx, other, op.ID, false); err == nil {
		t.Fatal("private v2 operation disclosed")
	}
	if page, err := w.writer.history(ctx, other, draft.ID, ""); err != nil || len(page.Items) != 0 {
		t.Fatal("foreign history disclosed", err)
	}
	if after, invalid := w.image.counts(); after != reads || invalid != bad {
		t.Fatal("local exclusion contacted upstream")
	}
	if err := attempt.Sent(ctx, op, writeback.Completion{Known: true}); err != nil {
		t.Fatal(err)
	}
	if err := attempt.Verify(ctx, op); err != nil {
		t.Fatal(err)
	}
	current, err = w.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || current.Status != "partial" {
		t.Fatal("known partial did not settle", err)
	}
	remaining, err := w.writer.confirm(ctx, other, input)
	if err != nil || remaining.Plan.Description != nil || len(remaining.Plan.Fields) != 1 || remaining.Plan.Fields[0] != "gps" {
		t.Fatal("settled remaining-field approval", err)
	}
}
