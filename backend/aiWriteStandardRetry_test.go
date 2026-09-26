package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStandardRetryRequiresAllBaselineAndRecoversAcceptedGenerationLocally(t *testing.T) {
	w := standardWriteFixture(t, []string{"description"})
	ctx := context.Background()
	op, err := w.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "standard-retry"})
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
	if err = attempt.Sent(ctx, op, writeback.Completion{Known: true}); err != nil {
		t.Fatal(err)
	}
	if err = attempt.Verify(ctx, op); err != nil {
		t.Fatal(err)
	}
	op, err = w.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || op.Status != "retryable" {
		t.Fatal("known all-baseline readback", err, op.Status)
	}
	w.meta["exifInfo"] = map[string]any{"latitude": 30, "longitude": 40, "description": w.preview.Plan.Description.Before.Value}
	accepted, err := w.writer.retry(ctx, testUserID, op.ID, op.Generation)
	if err != nil || accepted.Status != "queued" || accepted.Generation != op.Generation+1 {
		t.Fatal("exact selected baseline retry unavailable", err)
	}
	w.f.reopen(t)
	w.writer.drafts.results.jobs = w.f.store
	w.writer.capabilities = writeback.CapabilityPolicy{}
	w.f.now = w.f.now.Add(6 * time.Minute)
	reads, bad := w.image.counts()
	replayed, err := w.writer.retry(ctx, testUserID, op.ID, op.Generation)
	if err != nil || replayed.Generation != accepted.Generation || replayed.ID != accepted.ID {
		t.Fatal("accepted retry identity lost", err)
	}
	if next, invalid := w.image.counts(); next != reads || invalid != bad {
		t.Fatal("accepted retry replay contacted Immich")
	}
}
