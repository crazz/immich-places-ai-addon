package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStackRecoveredGenerationFencesEveryLateTargetWorker(t *testing.T) {
	for _, phase := range []string{"reserved", "sent", "verifying"} {
		t.Run(phase, func(t *testing.T) {
			f := stackWriteFixture(t)
			op := approveStackFixture(t, f, f.members)
			m := stackMutationServer(t, f)
			ctx := context.Background()
			asset := op.Targets[0].AssetID
			step, err := writeback.TargetStep(op, asset)
			if err != nil {
				t.Fatal(err)
			}
			old := &aiStackAttempt{store: f.writer, parent: op, assetID: asset}
			if _, err := old.Read(ctx, step); err != nil {
				t.Fatal(err)
			}
			if err := old.Reserve(ctx, step, false); err != nil {
				t.Fatal(err)
			}
			outcome := writeback.Completion{}
			if phase != "reserved" {
				outcome = old.Send(ctx, step)
			}
			if phase == "verifying" {
				if err := old.Sent(ctx, step, outcome); err != nil {
					t.Fatal(err)
				}
			}
			f.f.reopen(t)
			f.writer.drafts.results.jobs = f.f.store
			f.f.now = f.f.now.Add(time.Minute)
			if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
				t.Fatal(err)
			}
			saved, err := f.writer.get(ctx, testUserID, op.ID, false)
			wantStatus, wantSends := "succeeded", 1
			if phase == "reserved" {
				wantStatus, wantSends = "verifying", 0
			}
			if err != nil || saved.Targets[0].Status != wantStatus || saved.Targets[0].Attempts != 1 || saved.Targets[0].Generation != 2 {
				t.Fatal("incorrect independent recovery", err)
			}
			if err := old.Sent(ctx, step, writeback.Completion{Known: true, Code: "late"}); err == nil {
				t.Fatal("stale target worker published completion")
			}
			if err := old.Verify(ctx, step); err == nil {
				t.Fatal("stale target worker published readback")
			}
			if outcome := old.Send(ctx, step); !outcome.Known || outcome.Code != "not_sent" || m.sent(asset) != wantSends {
				t.Fatal("stale target worker resent mutation")
			}
			after, err := f.writer.get(ctx, testUserID, op.ID, false)
			if err != nil || after.Targets[0].Status != wantStatus || after.Targets[0].Generation != 2 || after.Targets[0].Code != saved.Targets[0].Code {
				t.Fatal("late worker changed recovered state", err)
			}
			for _, target := range after.Targets[1:] {
				if target.Generation != 0 || target.Attempts != 0 || target.Status != "queued" || m.sent(target.AssetID) != 0 {
					t.Fatal("target recovery changed sibling authority")
				}
			}
		})
	}
}
