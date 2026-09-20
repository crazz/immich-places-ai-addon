package main

import (
	"context"
	"reflect"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIJobSubmissionSurvivesReopenWithExactMembership(t *testing.T) {
	ctx := context.Background()
	directory := t.TempDir()
	db, err := newDatabase(directory, "job-test-key")
	if err != nil {
		t.Fatal(err)
	}
	if err = db.createUser(ctx, testUserID, "job@example.com", "hash"); err != nil {
		t.Fatal(err)
	}
	selection := selectionStore(t, db)
	if _, err = db.createAIProvider(ctx, testUserID, "profile", providerInput()); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	store := newAIJobStore(db, selection.binding, true, func() time.Time { return now })
	input := jobs.Submission{Owner: testUserID, Installation: selection.binding, Key: "submission-1", Profile: "profile", Revision: 1, AssetIDs: []string{selectionA, selectionB, selectionA}, Languages: []string{"UK", "en"}, PrimaryLanguage: "uk", SelectionDigest: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ConsentVersion: "visual-v1", MaxCalls: 6}
	job, err := store.Submit(ctx, input)
	if err != nil || job.ID == "" {
		t.Fatal("submit", err)
	}
	db.close()
	db, err = newDatabase(directory, "job-test-key")
	if err != nil {
		t.Fatal(err)
	}
	defer db.close()
	store = newAIJobStore(db, selection.binding, true, func() time.Time { return now })
	loaded, err := store.Get(ctx, testUserID, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Model != "manual-model" || loaded.Input.Revision != 1 || loaded.Input.MaxCalls != 6 || loaded.Input.ConsentVersion != "visual-v1" || !reflect.DeepEqual(loaded.Input.Languages, []string{"uk", "en"}) || !reflect.DeepEqual(loaded.Input.AssetIDs, []string{selectionA, selectionB}) || len(loaded.Items) != 2 {
		t.Fatalf("binding lost: %+v", loaded)
	}
	for _, item := range loaded.Items {
		if item.State != "queued" || item.Attempts != 0 || item.Calls != 0 {
			t.Fatalf("not queued: %+v", item)
		}
	}
}
