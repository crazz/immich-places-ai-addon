package main

import (
	"context"
	"errors"
	"immich-places-backend/internal/ai/jobs"
	"testing"
)

func TestAIResearchIdempotencyBindsExactHintWithoutAdditionalWork(t *testing.T) {
	f, p, req := researchProductionFixture(t)
	ctx := context.Background()
	first, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	again, err := p.submit(ctx, testUserID, req)
	if err != nil || again.ID != first.ID {
		t.Fatal("Research submission duplicated", err)
	}
	req.Configuration.Context.Hint = "Different place"
	req.Consent.Configuration = req.Configuration
	if _, err = p.submit(ctx, testUserID, req); !errors.Is(err, jobs.ErrConflict) {
		t.Fatal("changed hint reused existing submission", err)
	}
	if f.hits.Load() != 0 {
		t.Fatal("admission sent a photo")
	}
}
