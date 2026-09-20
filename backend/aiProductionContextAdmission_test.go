package main

import (
	"context"
	"encoding/json"
	"immich-places-backend/internal/ai/jobs"
	"testing"
)

func TestAIProductionAdmitsOnlyExplicitContextPolicy(t *testing.T) {
	f, p, req := contextProductionFixture(t)
	ctx := context.Background()
	before, _ := f.image.counts()
	job, err := p.submit(ctx, testUserID, req)
	if err != nil || job.Input.Mode != "context-assisted" {
		t.Fatal("context admission unavailable", err)
	}
	var raw string
	if err = f.image.db.db.QueryRow("SELECT requestJSON FROM ai_job_admissions WHERE jobID=?", job.ID).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var retained jobs.Admission
	_ = json.Unmarshal([]byte(raw), &retained)
	if retained.Configuration.Context == nil || retained.Configuration.Context.Classes == nil {
		t.Fatal("empty context consent lost")
	}
	_, oldP, oldReq := productionFixture(t)
	oldReq.Configuration.Mode = "context-assisted"
	oldReq.Configuration.Context = req.Configuration.Context
	oldReq.Consent.Configuration = oldReq.Configuration
	if _, err = oldP.submit(ctx, testUserID, oldReq); err == nil {
		t.Fatal("Visual policy granted Context")
	}
	after, _ := f.image.counts()
	if before != after || f.hits.Load() != 0 {
		t.Fatal("admission called upstream")
	}
}
