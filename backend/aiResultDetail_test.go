package main

import (
	"context"
	"immich-places-backend/internal/ai/jobs"
	"testing"
)

func TestAIResultDetailSeparatesImmutableUnknownFromCanceledWork(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	lease, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok {
		t.Fatal(err)
	}
	if err = f.store.Reserve(ctx, lease); err != nil {
		t.Fatal(err)
	}
	id, err := f.store.Complete(ctx, lease, aiJobCompletion(t))
	if err != nil {
		t.Fatal(err)
	}
	if err = f.store.Cancel(ctx, testUserID, job.ID); err != nil {
		t.Fatal(err)
	}
	f.store.enabled = false
	s := &aiResultStore{jobs: f.store}
	detail, err := s.detail(ctx, testUserID, id, "", "")
	if err != nil || detail.Proposal == nil || detail.Proposal.Outcome != "unknown" || detail.Entry.ExecutionState != "succeeded" || detail.Provenance == nil || detail.Provenance.Model != "manual-model" || detail.Entry.ReviewState != "unreviewed" || detail.Entry.WriteState != "not_requested" {
		t.Fatal("immutable detail unavailable", err)
	}
	resolved, err := s.detail(ctx, testUserID, "", job.ID, lease.ItemID)
	if err != nil || resolved.Entry.AnalysisID == nil || *resolved.Entry.AnalysisID != id {
		t.Fatal("item detail did not resolve", err)
	}
	for _, item := range job.Items {
		if item.ID != lease.ItemID {
			failed, err := s.detail(ctx, testUserID, "", job.ID, item.ID)
			if err != nil || failed.Proposal != nil || failed.Entry.ProposalOutcome != nil || failed.Entry.ExecutionState != "canceled" || failed.Entry.Failure != "canceled" {
				t.Fatal("fabricated failed proposal", err)
			}
		}
	}
}
