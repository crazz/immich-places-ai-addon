package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionStorageFailuresNeverPublishOrInventAccounting(t *testing.T) {
	for _, phase := range []string{"reservation", "usage", "completion"} {
		t.Run(phase, func(t *testing.T) {
			f, p, req := productionFixture(t)
			ctx := context.Background()
			job, err := p.submit(ctx, testUserID, req)
			if err != nil {
				t.Fatal(err)
			}
			table, event := "ai_job_usage", "INSERT"
			if phase == "usage" {
				event = "UPDATE"
				f.handle = func(w http.ResponseWriter, r *http.Request) bool {
					_, _ = w.Write([]byte(`{"choices":[],"usage":{"prompt_tokens":10}}`))
					return true
				}
			}
			if phase == "completion" {
				table = "ai_analyses"
			}
			selectionSQL(t, f.image.db, "CREATE TRIGGER fail_publication BEFORE "+event+" ON "+table+" BEGIN SELECT RAISE(ABORT,'private SQL details'); END")
			worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
			if _, err = worker.RunOne(ctx, make(chan time.Time)); err != jobs.ErrStorage {
				t.Fatal("storage failure hidden", err)
			}
			progress, err := p.progress(ctx, testUserID, job.ID)
			if err != nil || progress.Items[0].ResultID != nil || progress.Counts["running"] != 1 {
				t.Fatal("partial publication", progress, err)
			}
			expected := 1
			if phase == "reservation" {
				expected = 0
			}
			if progress.Usage.Calls != expected || int(f.hits.Load()) != expected || progress.Usage.ReservedTokens != int64(expected)*104000 {
				t.Fatal("partial call/token reservation", progress.Usage, f.hits.Load())
			}
		})
	}
}
