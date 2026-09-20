package main

import (
	"context"
	"testing"
)

func TestAIProductionAdmissionStorageFailureRollsBackEveryRecord(t *testing.T) {
	for _, table := range []string{"ai_job_items", "ai_job_launch", "ai_job_admissions", "commit"} {
		t.Run(table, func(t *testing.T) {
			f, p, req := productionFixture(t)
			if table == "commit" {
				selectionSQL(t, f.image.db, "CREATE TABLE deferred_failure(owner TEXT REFERENCES users(ID) DEFERRABLE INITIALLY DEFERRED)")
				selectionSQL(t, f.image.db, "CREATE TRIGGER reject_admission AFTER INSERT ON ai_job_admissions BEGIN INSERT INTO deferred_failure VALUES('absent'); END")
			} else {
				selectionSQL(t, f.image.db, "CREATE TRIGGER reject_admission BEFORE INSERT ON "+table+" BEGIN SELECT RAISE(ABORT,'private database detail'); END")
			}
			if job, err := p.submit(context.Background(), testUserID, req); err == nil || job.ID != "" {
				t.Fatal("partial admission returned")
			}
			for _, name := range []string{"ai_jobs", "ai_job_items", "ai_job_launch", "ai_job_admissions"} {
				var n int
				if err := f.image.db.db.QueryRow("SELECT count(*) FROM " + name).Scan(&n); err != nil || n != 0 {
					t.Fatal("partial record", name, n, err)
				}
			}
			if f.hits.Load() != 0 {
				t.Fatal("failed admission contacted provider")
			}
		})
	}
}
