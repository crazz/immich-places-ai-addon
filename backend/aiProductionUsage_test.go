package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionReportedUsageNeverRefundsAllowance(t *testing.T) {
	f, p, req := productionFixture(t)
	ctx := context.Background()
	content, err := os.ReadFile("internal/ai/results/testdata/ai-analysis-result.unknown.json")
	if err != nil {
		t.Fatal(err)
	}
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": string(content)}}}, "usage": map[string]int{"prompt_tokens": 100, "completion_tokens": 50, "total_tokens": 150}})
		return true
	}
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
	if _, err = worker.RunOne(ctx, make(chan time.Time)); err != nil {
		t.Fatal(err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Usage.ReportedStatus != "complete" || progress.Usage.ReservedTokens != 104000 || progress.Usage.CostStatus != "unknown" {
		t.Fatal("reported usage/refund mismatch", progress.Usage, err)
	}
	var input, output int
	if err := f.image.db.db.QueryRow("SELECT inputReported,outputReported FROM ai_job_usage WHERE jobID=?", job.ID).Scan(&input, &output); err != nil || input != 100 || output != 50 {
		t.Fatal("reported usage missing", input, output, err)
	}
}
