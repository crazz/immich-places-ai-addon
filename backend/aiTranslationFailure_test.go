package main

import (
	"context"
	"net/http"
	"testing"
)

func TestAITranslationFailedReservationSettlesWithoutDispatchOrHiddenRetry(t *testing.T) {
	f, s, req := translationFixture(t)
	dispatcher := translationProvider(t, f, s, func(w http.ResponseWriter, r *http.Request) { t.Error("unreserved request dispatched") })
	req.ProfileRevision = 2
	req.Languages = []string{"en"}
	if _, err := s.submit(context.Background(), testUserID, req); err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.db, `CREATE TRIGGER reject_translation_usage BEFORE INSERT ON ai_translation_usage BEGIN SELECT RAISE(ABORT,'synthetic reservation failure'); END`)
	worked, err := s.runOne(context.Background(), dispatcher)
	if !worked || err != nil {
		t.Fatal("reservation failure left unsettled", worked, err)
	}
	run, err := s.submit(context.Background(), testUserID, req)
	if err != nil || run.Items[0].State != "failed" || run.Items[0].Text != nil {
		t.Fatal("failed reservation holds capacity", run, err)
	}
	if worked, err = s.runOne(context.Background(), dispatcher); worked || err != nil {
		t.Fatal("hidden retry", worked, err)
	}
}
