package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/contextual"
	"immich-places-backend/internal/ai/jobs"
)

func TestAIResearchRejectsExplicitContextRestriction(t *testing.T) {
	f, p, req := productionFixture(t)
	req.Configuration.Mode = "research"
	req.Configuration.Context = &jobs.ContextChoices{Version: "context-v1", Classes: []contextual.Class{contextual.Hint}, Hint: "Private hint"}
	req.Consent.Configuration = req.Configuration
	if _, err := p.submit(context.Background(), testUserID, req); err == nil {
		t.Fatal("Research bypassed explicit context restriction")
	}
	if f.hits.Load() != 0 {
		t.Fatal("denied submission dispatched")
	}
}

func TestAIResearchRejectsNeighborDisclosure(t *testing.T) {
	f, p, req := researchProductionFixture(t)
	req.Configuration.Context.Classes = append(req.Configuration.Context.Classes, contextual.Neighbors)
	req.Consent.Configuration = req.Configuration
	if _, err := p.submit(context.Background(), testUserID, req); err == nil {
		t.Fatal("Research admitted hidden neighboring metadata")
	}
	if f.hits.Load() != 0 {
		t.Fatal("denied submission dispatched")
	}
}
