package main

import (
	"context"
	"reflect"
	"slices"
	"testing"
)

func TestAIStackReviewPersistsExactSelectedObservations(t *testing.T) {
	f := stackWriteFixture(t)
	store := f.writer.drafts
	ctx := context.Background()
	selected := []string{f.draft.AssetID, selectionB, selectionB}
	review, err := store.observeStack(ctx, testUserID, f.draft.ID, f.draft.Revision, selected, f.image.service)
	if err != nil {
		t.Fatal(err)
	}
	if review.ID == "" || review.StackID != f.stackID || review.AnalyzedID != f.draft.AssetID || review.DraftRevision != f.draft.Revision || len(review.Targets) != 2 || len(review.Candidates) != 3 {
		t.Fatalf("incomplete or expanded review: %+v", review)
	}
	for _, target := range review.Targets {
		if !slices.Contains(selected, target.AssetID) || target.ImageIdentity == "" {
			t.Fatal("unapproved target")
		}
		if target.AssetID == selectionB && (target.Before.Latitude == nil || *target.Before.Latitude != 1 || target.Before.Longitude != nil) {
			t.Fatal("independent partial GPS lost")
		}
	}
	f.f.reopen(t)
	store.results.jobs = f.f.store
	before, _ := f.image.counts()
	saved, err := store.readStackReview(ctx, testUserID, f.draft.ID, f.draft.Revision, review.ID)
	if err != nil || !reflect.DeepEqual(saved, review) {
		t.Fatal("private review lost across reopen", err)
	}
	if _, err = store.readStackReview(ctx, "foreign", f.draft.ID, f.draft.Revision, review.ID); err == nil {
		t.Fatal("foreign owner read target baselines")
	}
	after, bad := f.image.counts()
	if before != after || bad != 0 {
		t.Fatal("reload used upstream or review mutated it", before, after, bad)
	}
}
