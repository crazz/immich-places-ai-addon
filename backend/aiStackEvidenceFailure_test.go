package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStackCatalogFailureRetainsVerifiedFieldsAndBlocksReplayAfterRevert(t *testing.T) {
	f := stackWriteFixture(t)
	op := approveStackFixture(t, f, f.members)
	ctx := context.Background()
	asset := op.Plan.TargetID
	step, err := writeback.TargetStep(op, asset)
	if err != nil {
		t.Fatal(err)
	}
	a := &aiStackAttempt{store: f.writer, parent: op, assetID: asset}
	if _, err := a.Read(ctx, step); err != nil {
		t.Fatal(err)
	}
	if err := a.Reserve(ctx, step, false); err != nil {
		t.Fatal(err)
	}
	if err := a.Sent(ctx, step, writeback.Completion{Known: true}); err != nil {
		t.Fatal(err)
	}
	f.metadata[asset]["exifInfo"] = map[string]any{"latitude": 0, "longitude": 12, "description": op.Plan.Description.Intended}
	selectionSQL(t, f.f.db, `CREATE TRIGGER fail_stack_refresh BEFORE UPDATE OF latitude ON assets BEGIN SELECT RAISE(ABORT,'synthetic refresh failure'); END`)
	if err := a.Verify(ctx, step); err == nil {
		t.Fatal("expected local publication failure")
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || saved.Targets[0].Status != "verifying" || saved.Targets[0].Refreshed || saved.Targets[0].Code != "LOCAL_REFRESH_PENDING" {
		t.Fatal("catalog failure not distinguished from remote success", err)
	}
	for _, field := range saved.Targets[0].Fields {
		if !field.WasVerified || field.Status != "verified" {
			t.Fatal("verified evidence lost on catalog rollback")
		}
	}
	selectionSQL(t, f.f.db, `DROP TRIGGER fail_stack_refresh`)
	f.metadata[asset]["exifInfo"] = map[string]any{"latitude": nil, "longitude": nil, "description": op.Plan.Description.Before.Value}
	if err := a.Verify(ctx, step); err != nil {
		t.Fatal(err)
	}
	saved, err = f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || saved.Targets[0].Status != "conflict" || saved.Targets[0].Attempts != 1 {
		t.Fatal("external revert renewed approved payload", err)
	}
	if _, err := f.writer.retryStackTarget(ctx, testUserID, op.ID, asset, saved.Targets[0].Generation); err == nil {
		t.Fatal("verified field became retryable")
	}
	for _, target := range saved.Targets[1:] {
		if target.Status != "queued" || target.Attempts != 0 {
			t.Fatal("one target publication changed sibling state")
		}
	}
}
