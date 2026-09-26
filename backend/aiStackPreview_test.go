package main

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestAIStackPreviewFreezesReviewedSubsetAndRetainsBytes(t *testing.T) {
	f := stackWriteFixture(t)
	ctx := context.Background()
	review, err := f.writer.drafts.observeStack(ctx, testUserID, f.draft.ID, f.draft.Revision, []string{f.draft.AssetID, selectionB}, f.image.service)
	if err != nil {
		t.Fatal(err)
	}
	session := &aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service}
	preview, err := session.createStack(ctx, testUserID, f.draft.ID, f.draft.Revision, review.ID)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Plan.Version != "stack-preview-v3" || preview.Plan.Manifest == nil || len(preview.Plan.Manifest.Targets) != 2 || preview.Plan.Manifest.ReviewID != review.ID || preview.Plan.Manifest.StackID != f.stackID {
		t.Fatal("incorrect frozen scope")
	}
	for _, target := range preview.Plan.Manifest.Targets {
		if target.AssetID != f.draft.AssetID && (target.AssetID != selectionB || target.Description != nil || !reflect.DeepEqual(target.Fields, []string{"gps"})) {
			t.Fatal("scope expanded")
		}
	}
	var payload []byte
	if err := f.f.db.db.QueryRow(`SELECT payload FROM ai_stack_write_previews WHERE id=?`, preview.Plan.ID).Scan(&payload); err != nil {
		t.Fatal(err)
	}
	want, err := json.Marshal(preview.Plan)
	if err != nil || string(want) != string(payload) {
		t.Fatal("canonical bytes changed", err)
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	before, _ := f.image.counts()
	retained, err := session.get(ctx, testUserID, preview.Plan.ID)
	if err != nil || !reflect.DeepEqual(preview, retained) {
		t.Fatal("v3 preview lost across reopen", err)
	}
	after, bad := f.image.counts()
	if before != after || bad != 0 {
		t.Fatal("preview reload used network or review mutated")
	}
}
