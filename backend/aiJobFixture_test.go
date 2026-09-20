package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

type aiJobFixture struct {
	db        *Database
	store     *aiJobStore
	selection *aiSelectionStore
	input     jobs.Submission
	now       time.Time
}

func newAIJobFixture(t *testing.T) *aiJobFixture {
	t.Helper()
	db := newTestDB(t)
	selected := selectionStore(t, db)
	if _, err := db.createAIProvider(context.Background(), testUserID, "profile", providerInput()); err != nil {
		t.Fatal(err)
	}
	f := &aiJobFixture{db: db, selection: selected, now: time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)}
	f.store = newAIJobStore(db, selected.binding, true, func() time.Time { return f.now })
	f.input = jobs.Submission{Owner: testUserID, Installation: selected.binding, Key: "key", Profile: "profile", Revision: 1, AssetIDs: []string{selectionA, selectionB}, Languages: []string{"en"}, PrimaryLanguage: "en", SelectionDigest: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ConsentVersion: "visual-v1", MaxCalls: 6}
	return f
}
