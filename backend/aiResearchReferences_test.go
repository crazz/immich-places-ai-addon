package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIResearchDetailKeepsCoordinatesWithoutExposingUnsafeURL(t *testing.T) {
	f, p, req := researchProductionFixture(t, "https://user:secret@example.org/photo")
	ctx := context.Background()
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
	if _, err := worker.RunOne(ctx, make(chan time.Time)); err != nil {
		t.Fatal(err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || progress.Items[0].ResultID == nil {
		t.Fatal("estimate discarded", err)
	}
	s := &aiResultStore{jobs: p.store}
	detail, err := s.detail(ctx, testUserID, *progress.Items[0].ResultID, "", "")
	if err != nil || detail.Proposal == nil || detail.Proposal.Candidates[0].CameraLocation == nil {
		t.Fatal("coordinates unavailable", err)
	}
	raw, err := json.Marshal(detail)
	if err != nil || strings.Contains(string(raw), "user:secret") || detail.Proposal.Sources[0].URL != "" {
		t.Fatal("unsafe destination exposed", err)
	}
	record, err := p.store.ReadAnalysis(ctx, testUserID, *progress.Items[0].ResultID)
	if err != nil || !strings.Contains(string(record.Payload), "user:secret") {
		t.Fatal("immutable stored answer rewritten", err)
	}
}
