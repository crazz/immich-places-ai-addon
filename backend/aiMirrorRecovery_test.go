package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorRecoversLocalPublicationFailureAfterRestartWithoutResending(t *testing.T) {
	f := mirrorWriteFixture(t)
	mirrorMutationServer(t, f)
	ctx := context.Background()
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "publication-recovery"})
	if err != nil {
		t.Fatal(err)
	}
	if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = f.f.db.db.Exec(`CREATE TRIGGER fail_mirror_publication BEFORE UPDATE ON ai_mirror_records BEGIN SELECT RAISE(ABORT,'synthetic publication failure'); END`); err != nil {
		t.Fatal(err)
	}
	if err = f.writer.runOne(ctx, testUserID, op.ID); err == nil {
		t.Fatal("publication failure not observed")
	}
	if f.standardSends != 1 || f.mirrorSends != 1 {
		t.Fatal("wrong attempts before recovery")
	}
	if _, err = f.f.db.db.Exec(`DROP TRIGGER fail_mirror_publication`); err != nil {
		t.Fatal(err)
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	f.f.now = f.f.now.Add(2 * time.Second)
	f.writer.enabled = func() bool { return false }
	if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	op, err = f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || !op.Mirror.Verified || op.Mirror.Status != "succeeded" || !op.Verified {
		t.Fatal("read-only disabled recovery did not publish", err)
	}
	if f.standardSends != 1 || f.mirrorSends != 1 {
		t.Fatal("recovery repeated a mutation")
	}
}
