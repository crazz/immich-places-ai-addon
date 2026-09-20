package main

import (
	"context"
	"testing"
)

func TestAIResultTerminalProjectionIsAtomicAndSurvivesReopen(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.store.Cancel(ctx, testUserID, job.ID); err != nil {
		t.Fatal(err)
	}
	f.reopen(t)
	var count, distinct int
	if err = f.db.db.QueryRow(`SELECT count(*),count(DISTINCT sequence) FROM ai_result_history WHERE userID=? AND jobID=? AND executionState='canceled' AND captureDay IS NULL AND albumID IS NULL`, testUserID, job.ID).Scan(&count, &distinct); err != nil {
		t.Fatal(err)
	}
	if count != 2 || distinct != 2 {
		t.Fatalf("terminal history lost: %d/%d", count, distinct)
	}
	if err = f.store.Cancel(ctx, testUserID, job.ID); err != nil {
		t.Fatal(err)
	}
	var after int
	if err = f.db.db.QueryRow(`SELECT count(*) FROM ai_result_history WHERE userID=? AND jobID=?`, testUserID, job.ID).Scan(&after); err != nil || after != 2 {
		t.Fatal("repeat cancellation changed history", after, err)
	}
}
