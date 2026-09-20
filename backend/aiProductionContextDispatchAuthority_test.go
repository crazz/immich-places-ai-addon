package main

import (
	"context"
	"immich-places-backend/internal/ai/jobs"
	"testing"
	"time"
)

func TestAIProductionContextRechecksAuthorityAfterResolution(t *testing.T) {
	for _, kind := range []string{"consent", "policy", "revision", "cancel"} {
		t.Run(kind, func(t *testing.T) {
			f, p, req := contextProductionFixture(t)
			ctx := context.Background()
			job, err := p.submit(ctx, testUserID, req)
			if err != nil {
				t.Fatal(err)
			}
			f.analyzer.dispatcher.AfterResolve = func(context.Context) {
				switch kind {
				case "consent":
					selectionSQL(t, f.image.db, `UPDATE ai_job_admissions SET requestJSON=json_set(requestJSON,'$.consent.image',0) WHERE jobID=?`, job.ID)
				case "policy":
					p.policies, _ = jobs.ParseExecutionPolicies("")
				case "revision":
					selectionSQL(t, f.image.db, "UPDATE ai_provider_profiles SET activeRevision=2")
				case "cancel":
					if err := p.store.Cancel(ctx, testUserID, job.ID); err != nil {
						t.Fatal(err)
					}
				}
			}
			worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
			if _, err = worker.RunOne(ctx, make(chan time.Time)); err != nil {
				t.Fatal(err)
			}
			var state string
			var calls, results int
			if err = f.image.db.db.QueryRow("SELECT state,calls FROM ai_job_items WHERE jobID=?", job.ID).Scan(&state, &calls); err != nil {
				t.Fatal(err)
			}
			if err = f.image.db.db.QueryRow("SELECT count(*) FROM ai_analyses WHERE jobID=?", job.ID).Scan(&results); err != nil {
				t.Fatal(err)
			}
			expected := "blocked"
			if kind == "cancel" {
				expected = "canceled"
			}
			if state != expected || calls != 0 || f.hits.Load() != 0 || results != 0 {
				t.Fatal("post-resolution authority bypass", state, calls, results)
			}
		})
	}
}
