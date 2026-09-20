package main

import (
	"context"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/selection"
	"testing"
)

func TestAIProductionRerunChecksParentMembershipAndFreshEligibility(t *testing.T) {
	for _, tc := range []struct {
		name, state, kind string
		allowed           bool
	}{
		{"failed retry", "failed", "retry-failed", true},
		{"canceled retry", "canceled", "retry-failed", false},
		{"blocked retry", "blocked", "retry-failed", false},
		{"running retry", "running", "retry-failed", false},
		{"canceled reanalysis", "canceled", "reanalysis", true},
		{"foreign parent", "failed", "retry-failed", false},
		{"nonmember", "failed", "reanalysis", false},
		{"eligibility changed", "failed", "retry-failed", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, p, req := productionFixture(t)
			ctx := context.Background()
			parent, err := p.submit(ctx, testUserID, req)
			if err != nil {
				t.Fatal(err)
			}
			selectionSQL(t, f.image.db, "UPDATE ai_job_items SET state=? WHERE jobID=?", tc.state, parent.ID)
			asset := selectionA
			if tc.name == "nonmember" {
				asset = selectionB
				seedAsset(t, f.image.db, asset, nil, nil, "2026-09-20")
			}
			snapshot, err := p.selections.preview(ctx, selection.Input{Mode: "explicit", AssetIDs: []string{asset}, Scope: &selection.Scope{View: "all"}}, testUserID)
			if err != nil {
				t.Fatal(err)
			}
			req.Configuration.SelectionToken = *snapshot.SnapshotID
			req.Configuration.Rerun = &jobs.RerunChoice{ParentJobID: parent.ID, Kind: tc.kind}
			if tc.name == "foreign parent" {
				req.Configuration.Rerun.ParentJobID = selectionID(888)
			}
			if tc.name == "eligibility changed" {
				selectionSQL(t, f.image.db, "UPDATE assets SET isHidden=1 WHERE immichID=?", asset)
			}
			req.Consent.Configuration = req.Configuration
			req.IdempotencyKey = "new-rerun"
			_, err = p.submit(ctx, testUserID, req)
			if (err == nil) != tc.allowed {
				t.Fatal("incorrect rerun authority", err)
			}
			var count int
			_ = f.image.db.db.QueryRow("SELECT count(*) FROM ai_jobs").Scan(&count)
			expected := 1
			if tc.allowed {
				expected = 2
			}
			if count != expected {
				t.Fatal("partial or extra job")
			}
		})
	}
}
