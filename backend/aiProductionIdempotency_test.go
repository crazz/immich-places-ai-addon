package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionIdempotentRetrySurvivesSnapshotExpiry(t *testing.T) {
	f, p, req := productionFixture(t)
	ctx := context.Background()
	first, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	p.selections.now = func() time.Time { return time.Now().Add(24 * time.Hour) }
	selectionSQL(t, f.image.db, "DELETE FROM ai_selection_snapshots")
	retry, err := p.submit(ctx, testUserID, req)
	if err != nil || retry.ID != first.ID {
		t.Fatal("accepted retry requires preview", err)
	}
	req.Configuration.Limits.MaxCalls = 2
	req.Consent.Configuration = req.Configuration
	if _, err := p.submit(ctx, testUserID, req); err != jobs.ErrConflict {
		t.Fatal("changed key did not conflict", err)
	}
	req.IdempotencyKey = "fresh"
	if _, err := p.submit(ctx, testUserID, req); err == nil {
		t.Fatal("new admission accepted expired preview")
	}
	p.store.enabled = false
	req.IdempotencyKey = "production-one"
	req.Configuration.Limits.MaxCalls = 1
	req.Consent.Configuration = req.Configuration
	if _, err := p.submit(ctx, testUserID, req); err != jobs.ErrDenied {
		t.Fatal("disabled admission allowed", err)
	}
}
