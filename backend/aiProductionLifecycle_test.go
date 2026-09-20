package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionLifecycleFencesInstallationAndCascadesAccountData(t *testing.T) {
	f, p, req := productionFixture(t)
	ctx := context.Background()
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	s := &aiProductionWorkerStore{aiJobStore: p.store, production: p}
	lease, _, err := s.Claim(ctx, jobs.DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Reserve(ctx, lease); err != nil {
		t.Fatal(err)
	}
	if err = p.selections.bind(ctx, f.image.server.URL, "rotated"); err != nil {
		t.Fatal(err)
	}
	if err = s.Reserve(ctx, lease); err != jobs.ErrDenied {
		t.Fatal("old installation reserved", err)
	}
	if _, err = p.progress(ctx, testUserID, job.ID); err != jobs.ErrDenied {
		t.Fatal("old installation readable", err)
	}
	var state string
	if err = f.image.db.db.QueryRow("SELECT state FROM ai_job_items WHERE jobID=?", job.ID).Scan(&state); err != nil || state != "canceled" {
		t.Fatal("old work not fenced", state, err)
	}
	selectionSQL(t, f.image.db, "INSERT INTO ai_execution_policy_violations VALUES(?,?,1)", testUserID, req.Configuration.PolicyID)
	selectionSQL(t, f.image.db, "DELETE FROM users WHERE ID=?", testUserID)
	for _, table := range []string{"ai_jobs", "ai_job_items", "ai_job_admissions", "ai_job_launch", "ai_job_usage", "ai_execution_policy_violations"} {
		var n int
		if err = f.image.db.db.QueryRow("SELECT count(*) FROM " + table).Scan(&n); err != nil || n != 0 {
			t.Fatal("private lifecycle data retained", table, n, err)
		}
	}
}
