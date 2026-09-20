package main

import (
	"context"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/review"
	"testing"
)

func TestAIResultFailedAttemptHasSafeDetailWithoutAProposal(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	if _, err := f.store.Submit(ctx, f.input); err != nil {
		t.Fatal(err)
	}
	lease, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok {
		t.Fatal(err)
	}
	if err = f.store.Fail(ctx, lease, jobs.Failure{Code: jobs.Permanent}, 0); err != nil {
		t.Fatal(err)
	}
	s := &aiResultStore{jobs: f.store}
	page, err := s.list(ctx, testUserID, review.Query{Limit: 30, State: "failed"})
	if err != nil || len(page.Items) != 1 || page.Items[0].ProposalOutcome != nil || page.Items[0].AnalysisID != nil || page.Items[0].Failure != "execution" {
		t.Fatal("failure promoted to proposal", err)
	}
	detail, err := s.detail(ctx, testUserID, "", lease.JobID, lease.ItemID)
	if err != nil || detail.Proposal != nil || detail.Provenance == nil || detail.Provenance.Model != "manual-model" {
		t.Fatal("safe failed detail unavailable", err)
	}
}
