package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestAITranslationCancelRetainsInFlightCapacityAndFencesLateOutput(t *testing.T) {
	f, s, req := translationFixture(t)
	entered, release := make(chan struct{}), make(chan struct{})
	dispatcher := translationProvider(t, f, s, func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": `{"language":"en","status":"complete","text":"Late private text."}`}}}})
	})
	req.ProfileRevision = 2
	ctx := context.Background()
	run, err := s.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := s.runOne(ctx, dispatcher); done <- err }()
	<-entered
	err = s.cancel(ctx, testUserID, run.ID)
	var active int
	countErr := f.db.db.QueryRow("SELECT count(*) FROM ai_translation_items WHERE state='reserved'").Scan(&active)
	close(release)
	executionErr := <-done
	if err != nil || countErr != nil || active != 1 || executionErr != nil {
		t.Fatal("cancel released active sender early", err, countErr, active, executionErr)
	}
	run, err = s.submit(ctx, testUserID, req)
	if err != nil || run.Items[0].State != "canceled" || run.Items[0].Text != nil || run.Items[1].State != "canceled" {
		t.Fatal("canceled output published or pending send retained", run, err)
	}
	if worked, err := s.runOne(ctx, dispatcher); err != nil || worked {
		t.Fatal("canceled run resent", worked, err)
	}
}

func TestAITranslationRestartRetainsInterruptedOutcomeWithoutResend(t *testing.T) {
	f, s, req := translationFixture(t)
	dispatcher := translationProvider(t, f, s, func(w http.ResponseWriter, r *http.Request) { t.Error("interrupted provider call resent") })
	req.ProfileRevision = 2
	req.Languages = []string{"en"}
	ctx := context.Background()
	run, err := s.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.db, "UPDATE ai_translation_items SET state='reserved' WHERE runID=?", run.ID)
	f.reopen(t)
	s.drafts.results.jobs = f.store
	if err = s.recover(ctx); err != nil {
		t.Fatal(err)
	}
	if worked, err := s.runOne(ctx, dispatcher); err != nil || worked {
		t.Fatal("restart dispatched", worked, err)
	}
	run, err = s.submit(ctx, testUserID, req)
	if err != nil || run.Items[0].State != "interrupted" || run.Items[0].Text != nil {
		t.Fatal("restart outcome missing", run, err)
	}
}
