package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStandardApprovalRetainsExactPlanAndLocalIdempotency(t *testing.T) {
	w := standardWriteFixture(t, []string{"description"})
	ctx := context.Background()
	input := writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "description-approval"}
	reads, bad := w.image.counts()
	op, err := w.writer.confirm(ctx, testUserID, input)
	if err != nil {
		t.Fatal(err)
	}
	if op.Status != "queued" || op.Digest != input.Digest || op.Plan.Description == nil || op.Plan.Description.Intended != w.preview.Plan.Description.Intended {
		t.Fatal("confirmation changed reviewed authority")
	}
	w.f.reopen(t)
	w.writer.drafts.results.jobs = w.f.store
	w.writer.capabilities = writeback.CapabilityPolicy{}
	w.writer.enabled = func() bool { return false }
	retained, err := w.writer.confirm(ctx, testUserID, input)
	if err != nil || retained.ID != op.ID || retained.Digest != op.Digest {
		t.Fatal("lost accepted identity after reopen/disablement", err)
	}
	if after, invalid := w.image.counts(); after != reads || invalid != bad {
		t.Fatal("confirmation performed upstream I/O")
	}
	var protected int
	if err := w.f.db.db.QueryRow(`SELECT protected FROM ai_standard_write_previews WHERE id=?`, input.PreviewID).Scan(&protected); err != nil || protected != 1 {
		t.Fatal("approval did not protect exact preview", err)
	}
	if _, err := w.writer.get(ctx, "foreign", op.ID, false); err == nil {
		t.Fatal("foreign owner accessed approval")
	}
}
