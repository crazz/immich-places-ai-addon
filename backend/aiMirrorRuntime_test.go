package main

import (
	"context"
	"path/filepath"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorRuntimeSchedulesOptionalStepAfterStandardSettles(t *testing.T) {
	f := mirrorWriteFixture(t)
	mirrorMutationServer(t, f)
	ctx := context.Background()
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "runtime-metadata"})
	if err != nil {
		t.Fatal(err)
	}
	lock, err := acquireAIWriteLock(filepath.Join(t.TempDir(), "mirror.lock"))
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	r := &aiWriteRuntime{store: f.writer, lock: lock}
	for range 2 {
		if err = r.sweep(ctx); err != nil {
			t.Fatal(err)
		}
	}
	op, err = f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || !op.Verified || f.standardSends != 1 || f.mirrorSends != 1 {
		t.Fatal("durable runtime missed selected metadata step", err)
	}
}
