package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStandardRecoveryUsesSameBoundedRuntimeAndNeverResends(t *testing.T) {
	w := standardWriteFixture(t, []string{"description"})
	ctx := context.Background()
	op, err := w.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "standard-recovery"})
	if err != nil {
		t.Fatal(err)
	}
	attempt := &aiWriteAttempt{store: w.writer}
	if _, err = attempt.Read(ctx, op); err != nil {
		t.Fatal(err)
	}
	if err = attempt.Reserve(ctx, op, false); err != nil {
		t.Fatal(err)
	}
	if err = attempt.Sent(ctx, op, writeback.Completion{Known: true, Code: "synthetic_received"}); err != nil {
		t.Fatal(err)
	}
	w.meta["exifInfo"] = map[string]any{"description": w.preview.Plan.Description.Intended}
	w.f.reopen(t)
	w.writer.drafts.results.jobs = w.f.store
	w.f.now = w.f.now.Add(46 * time.Second)
	lock, err := acquireAIWriteLock(t.TempDir() + "/writer.lock")
	if err != nil || lock == nil {
		t.Fatal("runtime lock", err)
	}
	t.Cleanup(func() { lock.Close() })
	runtime := &aiWriteRuntime{store: w.writer, lock: lock}
	reads, bad := w.image.counts()
	if err = runtime.sweep(ctx); err != nil {
		t.Fatal(err)
	}
	retained, err := w.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || retained.Status != "succeeded" || retained.Attempts != 1 || retained.Generation != 2 {
		t.Fatalf("v2 recovery: %s attempts=%d generation=%d %v", retained.Status, retained.Attempts, retained.Generation, err)
	}
	if next, invalid := w.image.counts(); next != reads+1 || invalid != bad {
		t.Fatal("recovery changed mutation count or skipped readback")
	}
	page, err := w.writer.history(ctx, testUserID, w.draft.ID, "")
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != op.ID {
		t.Fatal("v2 history missing", err)
	}
}
