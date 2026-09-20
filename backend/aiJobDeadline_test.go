package main

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIJobWriteQueriesReceiveBoundedTransactionContext(t *testing.T) {
	f := newAIJobFixture(t)
	job, err := f.store.Submit(context.Background(), f.input)
	if err != nil {
		t.Fatal(err)
	}
	err = f.store.write(context.Background(), func(ctx context.Context, tx *sql.Tx) error {
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > 5*time.Second || time.Until(deadline) <= 0 {
			t.Fatal("query context is not bounded")
		}
		if _, err := tx.ExecContext(ctx, "UPDATE ai_jobs SET calls=1 WHERE userID=? AND id=?", testUserID, job.ID); err != nil {
			return err
		}
		return jobs.ErrInvalid
	})
	if err != jobs.ErrInvalid {
		t.Fatal("safe rollback error", err)
	}
	loaded, err := f.store.Get(context.Background(), testUserID, job.ID)
	if err != nil || loaded.Calls != 0 {
		t.Fatal("transaction did not roll back", err)
	}
}
