package main

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStackConfirmationPersistsEveryExactTargetAtomically(t *testing.T) {
	f := stackWriteFixture(t)
	ctx := context.Background()
	review, err := f.writer.drafts.observeStack(ctx, testUserID, f.draft.ID, f.draft.Revision, f.members, f.image.service)
	if err != nil {
		t.Fatal(err)
	}
	preview, err := (&aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service}).createStack(ctx, testUserID, f.draft.ID, f.draft.Revision, review.ID)
	if err != nil {
		t.Fatal(err)
	}
	input := writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "exact-stack"}
	op, err := f.writer.confirm(ctx, testUserID, input)
	if err != nil {
		t.Fatal("approved manifest unavailable", err)
	}
	var view struct {
		Targets []struct {
			AssetID              string `json:"assetId"`
			Status               string `json:"status"`
			Attempts, Generation int
			Fields               []writeback.FieldOutcome `json:"fields"`
		}
	}
	raw, err := json.Marshal(op)
	if err != nil || json.Unmarshal(raw, &view) != nil || len(view.Targets) != len(preview.Plan.Manifest.Targets) {
		t.Fatal("missing independent targets", err)
	}
	for i, target := range view.Targets {
		want := preview.Plan.Manifest.Targets[i]
		if target.AssetID != want.AssetID || target.Status != "queued" || target.Attempts != 0 || target.Generation != 0 || len(target.Fields) != len(want.Fields) {
			t.Fatal("incorrect initial target state")
		}
		for j, field := range target.Fields {
			if field.Field != want.Fields[j] || field.Status != "pending" {
				t.Fatal("unapproved field")
			}
		}
	}
	var guards, tokens int
	if err := f.f.db.db.QueryRow(`SELECT count(*),count(DISTINCT token) FROM ai_write_target_guards WHERE installationID=?`, op.Plan.Installation).Scan(&guards, &tokens); err != nil || guards != 3 || tokens != 3 {
		t.Fatal("targets do not have independent exclusion", guards, tokens, err)
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	f.writer.enabled = func() bool { return false }
	before, writes := f.image.counts()
	replayed, err := f.writer.confirm(ctx, testUserID, input)
	if err != nil || !reflect.DeepEqual(replayed, op) {
		t.Fatal("durable confirmation replay changed", err)
	}
	if _, err = f.writer.get(ctx, "foreign", op.ID, false); err == nil {
		t.Fatal("foreign owner read approval")
	}
	after, laterWrites := f.image.counts()
	if before != after || writes != 0 || laterWrites != 0 {
		t.Fatal("confirmation or reload contacted mutation boundary")
	}
}
