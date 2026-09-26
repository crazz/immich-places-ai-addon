package main

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/writeback"
)

func TestAIStackAccountDeletionRetainsOnlyUnknownOpaqueExclusion(t *testing.T) {
	for _, scenario := range []string{"queued", "unknown", "late-known"} {
		t.Run(scenario, func(t *testing.T) {
			f := stackWriteFixture(t)
			op := approveStackFixture(t, f, f.members)
			ctx := context.Background()
			asset := op.Plan.Manifest.Targets[1].AssetID
			step, err := writeback.TargetStep(op, asset)
			if err != nil {
				t.Fatal(err)
			}
			a := &aiStackAttempt{store: f.writer, parent: op, assetID: asset}
			if scenario != "queued" {
				if _, err := a.Read(ctx, step); err != nil {
					t.Fatal(err)
				}
				if err := a.Reserve(ctx, step, false); err != nil {
					t.Fatal(err)
				}
			}
			otherToken := uuid.NewString()
			if _, err := f.f.db.db.Exec(`INSERT INTO ai_write_target_guards(installationID,assetID,token) VALUES(?,?,?)`, op.Plan.Installation, selectionID(99), otherToken); err != nil {
				t.Fatal(err)
			}
			if _, err := f.f.db.db.Exec(`DELETE FROM users WHERE ID=?`, testUserID); err != nil {
				t.Fatal(err)
			}
			if scenario == "late-known" {
				if err := a.Sent(ctx, step, writeback.Completion{Known: true}); err == nil {
					t.Fatal("deleted target was recreated")
				}
			}
			var guards int
			want := 1
			if scenario == "unknown" {
				want++
			}
			if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards`).Scan(&guards); err != nil || guards != want {
				t.Fatal("incorrect opaque guard retention", guards, want, err)
			}
			var retained string
			if err := f.f.db.db.QueryRow(`SELECT token FROM ai_write_target_guards WHERE assetID=?`, selectionID(99)).Scan(&retained); err != nil || retained != otherToken {
				t.Fatal("unrelated opaque guard changed", err)
			}
			for _, table := range []string{"ai_stack_write_operations", "ai_stack_write_targets", "ai_stack_write_events", "ai_stack_write_fields", "ai_stack_write_previews", "ai_stack_reviews"} {
				var count int
				if err := f.f.db.db.QueryRow(`SELECT count(*) FROM `+table+` WHERE userID=?`, testUserID).Scan(&count); err != nil || count != 0 {
					t.Fatal("deleted private state survived", table, count, err)
				}
			}
		})
	}
}
