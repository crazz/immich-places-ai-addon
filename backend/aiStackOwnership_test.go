package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIStackUnknownSenderExcludesOtherVersionsAndDeletionPreservesOtherOwner(t *testing.T) {
	for _, version := range []string{"gps-preview-v1", "standard-preview-v2"} {
		t.Run(version, func(t *testing.T) {
			f := stackWriteFixture(t)
			op := approveStackFixture(t, f, f.members)
			ctx := context.Background()
			step, err := writeback.TargetStep(op, op.Plan.TargetID)
			if err != nil {
				t.Fatal(err)
			}
			a := &aiStackAttempt{store: f.writer, parent: op, assetID: op.Plan.TargetID}
			if _, err := a.Read(ctx, step); err != nil {
				t.Fatal(err)
			}
			if err := a.Reserve(ctx, step, false); err != nil {
				t.Fatal(err)
			}
			analysis, err := f.f.store.ReadAnalysis(ctx, testUserID, f.draft.AnalysisID)
			if err != nil {
				t.Fatal(err)
			}
			if err := f.f.store.Cancel(ctx, testUserID, analysis.JobID); err != nil {
				t.Fatal(err)
			}
			other := "stack-overlap-owner"
			if err := f.f.db.createUser(ctx, other, "overlap@example.com", "hash"); err != nil {
				t.Fatal(err)
			}
			key := "private-immich-image-key"
			if err := f.f.db.updateImmichAPIKey(ctx, other, &key); err != nil {
				t.Fatal(err)
			}
			if err := f.f.db.upsertAssets(ctx, other, []AssetRow{{ImmichID: op.Plan.TargetID, Type: "IMAGE", OriginalFileName: "synthetic.jpg", FileCreatedAt: "2026-09-20"}}); err != nil {
				t.Fatal(err)
			}
			if _, err := f.f.db.createAIProvider(ctx, other, "profile", providerInput()); err != nil {
				t.Fatal(err)
			}
			f.f.input.Owner, f.f.input.AssetIDs, f.f.input.MaxCalls = other, []string{op.Plan.TargetID}, 3
			job := completeAIJob(t, f.f, "other-stack-owner")
			var analysisID string
			if err := f.f.db.db.QueryRow(`SELECT id FROM ai_analyses WHERE userID=? AND jobID=?`, other, job.ID).Scan(&analysisID); err != nil {
				t.Fatal(err)
			}
			store := f.writer.drafts
			draft, err := store.accept(ctx, other, analysisID, nil)
			if err != nil {
				t.Fatal(err)
			}
			observation, err := store.observe(ctx, other, draft.ID, draft.Revision, f.image.service)
			if err != nil {
				t.Fatal(err)
			}
			draft, err = store.acknowledge(ctx, other, draft.ID, draft.Revision, observation.ID, f.image.service)
			if err != nil {
				t.Fatal(err)
			}
			edit := drafts.Edit{Camera: json.RawMessage(`{"latitude":0,"longitude":12}`), Fields: json.RawMessage(`["gps"]`), State: "staged"}
			if version == "standard-preview-v2" {
				language, policy, text := "en", "replace", "Other owner's reviewed description"
				edit.Fields = json.RawMessage(`["gps","description"]`)
				edit.PrimaryLanguage, edit.DescriptionPolicy = &language, &policy
				edit.Descriptions = map[string]drafts.DescriptionEdit{"en": {Text: &text}}
			}
			draft, err = store.edit(ctx, other, draft.ID, draft.Revision, edit)
			if err != nil {
				t.Fatal(err)
			}
			session := &aiWritePreviewSession{drafts: store, images: f.image.service}
			preview, err := writepreview.Create(ctx, session, session, session, other, draft.ID, draft.Revision, uuid.NewString(), f.f.store.now)
			if err != nil || preview.Plan.Version != version {
				t.Fatal("foreign approval fixture", err)
			}
			input := writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "other-approved"}
			reads, bad := f.image.counts()
			if _, err := f.writer.confirm(ctx, other, input); err == nil || err.Error() != "TARGET_BUSY" {
				t.Fatal("v3 exclusion did not reject foreign version privately", err)
			}
			if _, err := f.writer.get(ctx, other, op.ID, false); err == nil {
				t.Fatal("foreign v3 private operation disclosed")
			}
			if page, err := f.writer.history(ctx, other, draft.ID, ""); err != nil || len(page.Items) != 0 {
				t.Fatal("foreign private history disclosed", err)
			}
			selectionSQL(t, f.f.db, `DELETE FROM users WHERE ID=?`, testUserID)
			var retained drafts.Draft
			if err := store.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
				var err error
				retained, err = store.read(ctx, tx, other, draft.ID)
				return err
			}); err != nil {
				t.Fatal("deletion removed unrelated owner", err)
			}
			before, _ := json.Marshal(draft)
			after, _ := json.Marshal(retained)
			if string(before) != string(after) {
				t.Fatal("unrelated private draft changed on owner deletion")
			}
			if _, err := f.writer.confirm(ctx, other, input); err == nil || err.Error() != "TARGET_BUSY" {
				t.Fatal("deletion discarded unknown sender exclusion", err)
			}
			if err := a.Sent(ctx, step, writeback.Completion{Known: true}); err == nil {
				t.Fatal("late sender recreated deleted private record")
			}
			remaining, err := f.writer.confirm(ctx, other, input)
			if err != nil || remaining.Plan.Version != version || remaining.Plan.Manifest != nil {
				t.Fatal("known-completion cleanup lost unrelated approval", err)
			}
			if after, invalid := f.image.counts(); after != reads || invalid != bad {
				t.Fatal("local ownership lifecycle contacted upstream")
			}
		})
	}
}
