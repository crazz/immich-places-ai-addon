package main

import (
	"context"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/review"
	"strings"
	"testing"
)

func TestAIResultRunsKeepOriginalLabelsAndFailedRunProvenance(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	f.input.AssetIDs = []string{selectionA}
	f.input.MaxCalls = 3
	if _, err := f.store.Submit(ctx, f.input); err != nil {
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
	input := providerInput()
	input.Model = "new-model"
	if _, err = f.db.updateAIProvider(ctx, testUserID, "profile", 1, input); err != nil {
		t.Fatal(err)
	}
	f.input.Key = "second"
	f.input.Revision = 2
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.store.Cancel(ctx, testUserID, job.ID); err != nil {
		t.Fatal(err)
	}
	s := &aiResultStore{jobs: f.store}
	page, err := s.list(ctx, testUserID, review.Query{Limit: 30})
	if err != nil || len(page.Items) != 2 {
		t.Fatal(err)
	}
	for _, entry := range page.Items {
		if entry.AnalysisID != nil {
			if entry.Label == nil || !strings.HasPrefix(*entry.Label, "A scene without") || len(*entry.Label) > 800 {
				t.Fatal("missing bounded primary label")
			}
		}
	}
	old, err := s.detail(ctx, testUserID, id, "", "")
	if err != nil || old.Provenance.Model != "manual-model" || old.Provenance.Revision != 1 {
		t.Fatal("old provenance replaced", err)
	}
	failed, err := s.detail(ctx, testUserID, "", job.ID, job.Items[0].ID)
	if err != nil || failed.Provenance == nil || failed.Provenance.Model != "new-model" || failed.Provenance.Revision != 2 || failed.Provenance.Profile != "profile" {
		t.Fatal("failed run lost original provenance", err)
	}
}
