package main

import (
	"context"
	"encoding/json"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/selection"
	"testing"
)

func TestAIProductionRerunCreatesExactFreshLinkedJob(t *testing.T) {
	f, p, req := productionFixture(t)
	ctx := context.Background()
	parent, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.image.db, "UPDATE ai_job_items SET state='failed',failure='execution' WHERE jobID=?", parent.ID)
	snapshot, err := p.selections.preview(ctx, selection.Input{Mode: "explicit", AssetIDs: []string{selectionA}, Scope: &selection.Scope{View: "all"}}, testUserID)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(req.Configuration)
	var cfg map[string]any
	_ = json.Unmarshal(raw, &cfg)
	cfg["selectionToken"] = *snapshot.SnapshotID
	cfg["rerun"] = map[string]any{"parentJobId": parent.ID, "kind": "retry-failed"}
	raw, _ = json.Marshal(map[string]any{"configuration": cfg, "consent": map[string]any{"version": "image-consent-v1", "image": true, "configuration": cfg}, "idempotencyKey": "fresh-rerun"})
	var rerun jobs.Admission
	if jobs.DecodeRequest(raw, &rerun) != nil {
		t.Fatal("rerun request unavailable")
	}
	child, err := p.submit(ctx, testUserID, rerun)
	if err != nil || child.ID == parent.ID {
		t.Fatal("new linked run unavailable", err)
	}
	original, err := p.store.Get(ctx, testUserID, parent.ID)
	if err != nil || original.Items[0].State != "failed" {
		t.Fatal("parent reset")
	}
	var retained string
	_ = f.image.db.db.QueryRow("SELECT requestJSON FROM ai_job_admissions WHERE jobID=?", child.ID).Scan(&retained)
	var stored map[string]any
	_ = json.Unmarshal([]byte(retained), &stored)
	linkage := stored["configuration"].(map[string]any)["rerun"].(map[string]any)
	if linkage["parentJobId"] != parent.ID || linkage["kind"] != "retry-failed" {
		t.Fatal("lineage missing")
	}
}
