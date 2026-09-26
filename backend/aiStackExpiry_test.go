package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStackExpiryStopsOnlyUnstartedTargetsAndRetainsReadback(t *testing.T) {
	f := stackWriteFixture(t)
	op := approveStackFixture(t, f, f.members)
	m := stackMutationServer(t, f)
	ctx := context.Background()
	if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	second := op.Plan.Manifest.Targets[1].AssetID
	step, err := writeback.TargetStep(op, second)
	if err != nil {
		t.Fatal(err)
	}
	a := &aiStackAttempt{store: f.writer, parent: op, assetID: second}
	if _, err := a.Read(ctx, step); err != nil {
		t.Fatal(err)
	}
	if err := a.Reserve(ctx, step, false); err != nil {
		t.Fatal(err)
	}
	if result := a.Send(ctx, step); result.Code != "response_received" {
		t.Fatal(result.Code)
	}
	if err := a.Sent(ctx, step, writeback.Completion{Known: false}); err != nil {
		t.Fatal(err)
	}
	f.f.now = f.f.now.Add(6 * time.Minute)
	for range 3 {
		if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
			t.Fatal(err)
		}
	}
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || saved.Status != "partial" || saved.Targets[0].Status != "succeeded" || saved.Targets[1].Status != "succeeded" || saved.Targets[2].Status != "expired" || saved.Targets[2].Code != "APPROVAL_EXPIRED" {
		t.Fatal("expiry lost honest independent outcomes", saved.Status, saved.Targets, err)
	}
	if saved.Targets[2].Attempts != 0 || m.sent(saved.Targets[2].AssetID) != 0 || m.sent(second) != 1 {
		t.Fatal("expiry started or repeated a mutation")
	}
}
