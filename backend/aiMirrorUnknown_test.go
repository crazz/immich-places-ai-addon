package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorUnknownSenderStaysUnsettledAndRecoveryNeverResends(t *testing.T) {
	f := mirrorWriteFixture(t)
	mirrorMutationServer(t, f)
	ctx := context.Background()
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "lost-sender"})
	if err != nil {
		t.Fatal(err)
	}
	if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	op, err = f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	a := &aiMirrorAttempt{standard: &aiStackAttempt{store: f.writer, parent: op, assetID: f.draft.AssetID}}
	if _, err = a.read(ctx, false); err != nil {
		t.Fatal(err)
	}
	if err = a.reserve(ctx, false); err != nil {
		t.Fatal(err)
	}
	if result := a.send(ctx); !result.Known {
		t.Fatal("fixture send failed")
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	f.writer.enabled = func() bool { return false }
	for range 4 {
		f.f.now = f.f.now.Add(46 * time.Second)
		if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
			t.Fatal(err)
		}
	}
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if !saved.Mirror.Verified || saved.Mirror.Settled || saved.Status != "verifying" || saved.Mirror.Attempts != 1 {
		t.Fatal("observed state fabricated sender completion", saved.Mirror)
	}
	if _, err = f.writer.retryMirror(ctx, testUserID, op.ID, f.draft.AssetID, saved.Mirror.Generation); err == nil {
		t.Fatal("unknown sender allowed retry")
	}
	var reads int
	if err = f.f.db.db.QueryRow(`SELECT reads FROM ai_mirror_write_steps WHERE operationID=?`, op.ID).Scan(&reads); err != nil || reads != 3 {
		t.Fatal("readback budget not bounded", reads, err)
	}
	if _, err = f.writer.reconcileMirror(ctx, testUserID, op.ID, f.draft.AssetID, saved.Mirror.Generation); err != nil {
		t.Fatal(err)
	}
	f.f.now = f.f.now.Add(2 * time.Second)
	if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	if f.standardSends != 1 || f.mirrorSends != 1 {
		t.Fatal("read-only recovery repeated a mutation")
	}
}
