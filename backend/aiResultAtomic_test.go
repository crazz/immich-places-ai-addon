package main

import (
	"context"
	"testing"
)

func TestAIResultProjectionFailureRollsBackTerminalState(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.db, `CREATE TRIGGER reject_history BEFORE INSERT ON ai_result_history BEGIN SELECT RAISE(ABORT,'synthetic history failure'); END`)
	if err = f.store.Cancel(ctx, testUserID, job.ID); err == nil {
		t.Fatal("terminal transition ignored history failure")
	}
	stored, err := f.store.Get(ctx, testUserID, job.ID)
	if err != nil || stored.Canceled {
		t.Fatal("partial cancellation survived", err)
	}
	for _, item := range stored.Items {
		if item.State != "queued" {
			t.Fatal("partial terminal item survived")
		}
	}
	var count int
	if err = f.db.db.QueryRow(`SELECT count(*) FROM ai_result_history`).Scan(&count); err != nil || count != 0 {
		t.Fatal("partial history survived", count, err)
	}
}
