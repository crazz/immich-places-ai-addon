package main

import (
	"context"
	"testing"
)

func TestAIJobInsertionFailureRollsBackHeaderAndItems(t *testing.T) {
	for _, failure := range []string{"item", "commit", "cancel"} {
		t.Run(failure, func(t *testing.T) {
			f := newAIJobFixture(t)
			switch failure {
			case "item":
				selectionSQL(t, f.db, `CREATE TRIGGER reject_job_item BEFORE INSERT ON ai_job_items WHEN NEW.position=1 BEGIN SELECT RAISE(ABORT,'private failure'); END`)
			case "commit":
				selectionSQL(t, f.db, `CREATE TABLE job_commit_failure(owner TEXT REFERENCES users(ID) DEFERRABLE INITIALLY DEFERRED)`)
				selectionSQL(t, f.db, `CREATE TRIGGER reject_job_commit AFTER INSERT ON ai_jobs BEGIN INSERT INTO job_commit_failure VALUES('foreign'); END`)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if failure == "cancel" {
				cancel()
			}
			if job, err := f.store.Submit(ctx, f.input); err == nil || job.ID != "" {
				t.Fatal("partial success")
			}
			for _, table := range []string{"ai_jobs", "ai_job_items"} {
				var count int
				if err := f.db.db.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 0 {
					t.Fatal("partial row", table, count, err)
				}
			}
		})
	}
}
