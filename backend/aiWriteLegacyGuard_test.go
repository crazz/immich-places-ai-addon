package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIStandardNewDescriptionCannotOverlapRetainedUnknownV1Owner(t *testing.T) {
	w := newAIWriteFixture(t)
	ctx := context.Background()
	attempt := &aiWriteAttempt{store: w.writer}
	if _, err := attempt.Read(ctx, w.op); err != nil {
		t.Fatal(err)
	}
	if err := attempt.Reserve(ctx, w.op, false); err != nil {
		t.Fatal(err)
	}
	if err := goose.DownTo(w.f.db.db, "migrations", 28); err != nil {
		t.Fatal(err)
	}
	if err := runMigrations(w.f.db.db); err != nil {
		t.Fatal(err)
	}
	w.f.reopen(t)
	w.writer.drafts.results.jobs = w.f.store
	w.writer.capabilities = writeback.CapabilityPolicy{Version: 1, Installation: w.f.store.binding, Profile: w.writer.profile, Evidence: "synthetic-new-version", Capabilities: []string{"description"}}
	other := "new-description-owner"
	analysis, err := w.f.store.ReadAnalysis(ctx, testUserID, w.draft.AnalysisID)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.f.store.Cancel(ctx, testUserID, analysis.JobID); err != nil {
		t.Fatal(err)
	}
	if err := w.f.db.createUser(ctx, other, "new-description@example.com", "hash"); err != nil {
		t.Fatal(err)
	}
	key := "private-immich-image-key"
	if err := w.f.db.updateImmichAPIKey(ctx, other, &key); err != nil {
		t.Fatal(err)
	}
	if err := w.f.db.upsertAssets(ctx, other, []AssetRow{{ImmichID: w.op.Plan.TargetID, Type: "IMAGE", OriginalFileName: "synthetic.jpg", FileCreatedAt: "2026-09-20"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.f.db.createAIProvider(ctx, other, "profile", providerInput()); err != nil {
		t.Fatal(err)
	}
	w.f.input.Owner, w.f.input.AssetIDs, w.f.input.MaxCalls = other, []string{w.op.Plan.TargetID}, 3
	job := completeAIJob(t, w.f, "new-description")
	var analysisID string
	if err := w.f.db.db.QueryRow("SELECT id FROM ai_analyses WHERE userID=? AND jobID=?", other, job.ID).Scan(&analysisID); err != nil {
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
	language, policy, text := "en", "replace", "An exact new description"
	draft, err = store.edit(ctx, other, draft.ID, draft.Revision, drafts.Edit{PrimaryLanguage: &language, DescriptionPolicy: &policy, Descriptions: map[string]drafts.DescriptionEdit{language: {Text: &text}}, Fields: json.RawMessage(`["description"]`), State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	session := &aiWritePreviewSession{drafts: store, images: w.image.service}
	preview, err := writepreview.Create(ctx, session, session, session, other, draft.ID, draft.Revision, uuid.NewString(), w.f.store.now)
	if err != nil || preview.Plan.Version != "standard-preview-v2" {
		t.Fatal("new standard preview", err)
	}
	sends, reads := w.counts()
	if _, err := w.writer.confirm(ctx, other, writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "overlap"}); err == nil || err.Error() != "TARGET_BUSY" {
		t.Fatal("old unresolved target did not block new owner/version", err)
	}
	if _, err := w.writer.get(ctx, other, w.op.ID, false); err == nil {
		t.Fatal("legacy private operation disclosed")
	}
	if afterSends, afterReads := w.counts(); afterSends != sends || afterReads != reads {
		t.Fatal("overlap check performed upstream I/O")
	}
	if old := w.status(t); old.Status != "writing" || old.Attempts != 1 || old.Settled {
		t.Fatal("new request changed old recovery state", old.Status)
	}
}
