package main

import (
	"context"
	"reflect"
	"slices"
	"testing"
)

func addStackFixtureMember(t *testing.T, f *aiStackFixture, id string) {
	t.Helper()
	seedAsset(t, f.f.db, id, nil, nil, "2026-09-20")
	if _, err := f.f.db.db.Exec(`UPDATE assets SET stackID=?,stackPrimaryAssetID=? WHERE userID=? AND immichID=?`, f.stackID, f.draft.AssetID, testUserID, id); err != nil {
		t.Fatal(err)
	}
	meta := f.image.metadata()
	meta["id"] = id
	meta["exifInfo"] = map[string]any{"latitude": -12, "longitude": 0, "description": "unselected text"}
	f.metadata[id] = meta
	f.members = append(f.members, id)
	for _, value := range f.metadata {
		value["stack"] = map[string]any{"id": f.stackID, "primaryAssetId": f.draft.AssetID, "assetCount": len(f.members)}
	}
}

func TestAIStackSelectsThreeReviewedPhotosFromFive(t *testing.T) {
	f := stackWriteFixture(t)
	addStackFixtureMember(t, f, selectionID(4))
	addStackFixtureMember(t, f, selectionID(5))
	selected := []string{f.draft.AssetID, selectionB, selectionID(4)}
	ctx := context.Background()
	review, err := f.writer.drafts.observeStack(ctx, testUserID, f.draft.ID, f.draft.Revision, selected, f.image.service)
	if err != nil || len(review.Candidates) != 5 {
		t.Fatal("five-photo review failed", err)
	}
	preview, err := (&aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service}).createStack(ctx, testUserID, f.draft.ID, f.draft.Revision, review.ID)
	if err != nil || preview.Plan.Manifest == nil || len(preview.Plan.Manifest.Targets) != 3 {
		t.Fatal("subset failed", err)
	}
	var got []string
	for _, target := range preview.Plan.Manifest.Targets {
		got = append(got, target.AssetID)
		if target.Intended.Latitude != 0 || target.Intended.Longitude != 12 {
			t.Fatal("intended camera changed")
		}
		if target.AssetID == selectionID(4) && (target.Before.Latitude == nil || *target.Before.Latitude != -12 || target.Before.Longitude == nil || *target.Before.Longitude != 0) {
			t.Fatal("independent reviewed baseline lost")
		}
	}
	slices.Sort(selected)
	if !reflect.DeepEqual(got, selected) {
		t.Fatalf("wrong targets %v", got)
	}
}

func TestAIStackDefaultPreviewStaysSinglePhotoAfterMemberAdded(t *testing.T) {
	f := stackWriteFixture(t)
	ctx := context.Background()
	review, err := f.writer.drafts.observeStack(ctx, testUserID, f.draft.ID, f.draft.Revision, nil, f.image.service)
	if err != nil || len(review.Targets) != 1 || review.Targets[0].AssetID != f.draft.AssetID {
		t.Fatal("default review expanded", err)
	}
	session := &aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service}
	preview, err := session.createStack(ctx, testUserID, f.draft.ID, f.draft.Revision, review.ID)
	if err != nil || preview.Plan.Manifest != nil || preview.Plan.TargetID != f.draft.AssetID {
		t.Fatal("default preview expanded", err)
	}
	addStackFixtureMember(t, f, selectionID(4))
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	before, _ := f.image.counts()
	saved, err := session.get(ctx, testUserID, preview.Plan.ID)
	if err != nil || !reflect.DeepEqual(saved, preview) || saved.Plan.Manifest != nil {
		t.Fatal("default scope changed after addition or reopen", err)
	}
	after, writes := f.image.counts()
	if before != after || writes != 0 {
		t.Fatal("local reload contacted Immich")
	}
}

func TestAIStackUnselectedAdditionPreservesReviewedAndSavedScope(t *testing.T) {
	f := stackWriteFixture(t)
	ctx := context.Background()
	review, err := f.writer.drafts.observeStack(ctx, testUserID, f.draft.ID, f.draft.Revision, []string{f.draft.AssetID, selectionB}, f.image.service)
	if err != nil {
		t.Fatal(err)
	}
	addStackFixtureMember(t, f, selectionID(4))
	session := &aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service}
	preview, err := session.createStack(ctx, testUserID, f.draft.ID, f.draft.Revision, review.ID)
	if err != nil || preview.Plan.Manifest == nil || len(preview.Plan.Manifest.Targets) != 2 {
		t.Fatal("unselected addition blocked or expanded reviewed scope", err)
	}
	for i, target := range preview.Plan.Manifest.Targets {
		if target.AssetID != review.Targets[i].AssetID || !reflect.DeepEqual(target.Before, review.Targets[i].Before) || target.ImageIdentity != review.Targets[i].ImageIdentity {
			t.Fatal("reviewed identity or comparison changed")
		}
	}
	addStackFixtureMember(t, f, selectionID(5))
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	before, _ := f.image.counts()
	saved, err := session.get(ctx, testUserID, preview.Plan.ID)
	if err != nil || !reflect.DeepEqual(saved, preview) {
		t.Fatal("saved manifest or digest changed after addition", err)
	}
	after, writes := f.image.counts()
	if before != after || writes != 0 {
		t.Fatal("reload contacted Immich")
	}
}
