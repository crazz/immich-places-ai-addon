package main

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStackRestartReconcilesAmbiguousMemberWithoutResendingSiblings(t *testing.T) {
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
		t.Fatal("fixture did not send", result.Code)
	}
	if err := a.Sent(ctx, step, writeback.Completion{Known: false, Code: "response_lost"}); err != nil {
		t.Fatal(err)
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	f.f.now = f.f.now.Add(46 * time.Second)
	lock, err := acquireAIWriteLock(filepath.Join(t.TempDir(), "stack-runtime.lock"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = lock.Close() })
	runtime := &aiWriteRuntime{store: f.writer, lock: lock}
	for range 4 {
		if err := runtime.sweep(ctx); err != nil {
			t.Fatal(err)
		}
		f.f.now = f.f.now.Add(time.Second)
	}
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || saved.Status != "succeeded" || !saved.Verified {
		t.Fatal("restart did not independently recover targets", saved.Status, err)
	}
	for _, target := range saved.Targets {
		if target.Attempts != 1 || m.sent(target.AssetID) != 1 {
			t.Fatal("missing or resent target", target.AssetID)
		}
	}
	if saved.Targets[1].Settled {
		t.Fatal("desired state invented prior sender completion")
	}
}
