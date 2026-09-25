package main

import (
	"context"
	"testing"
	"time"
)

func TestAIWriteSyncDrainFailureRemainsPendingAndRecoversWithoutResend(t *testing.T) {
	w := newAIWriteFixture(t)
	if !w.writer.sync.acquireUserSyncLock(testUserID) {
		t.Fatal("fixture lock")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_ = w.writer.runOne(ctx, testUserID, w.op.ID)
	op := w.status(t)
	if op.Status != "verifying" || op.Code != "LOCAL_REFRESH_PENDING" || op.Refreshed {
		t.Fatal("false refresh after drain failure", op.Status, op.Code)
	}
	w.writer.sync.releaseUserSyncLock(testUserID)
	w.f.now = w.f.now.Add(time.Second)
	w.run(t)
	if op = w.status(t); !op.Refreshed || op.Status != "succeeded" {
		t.Fatal(op)
	}
	if sends, _ := w.counts(); sends != 1 {
		t.Fatal("resent after drain failure", sends)
	}
	if !w.writer.sync.acquireUserSyncLock(testUserID) {
		t.Fatal("publication leaked sync lock")
	}
	w.writer.sync.releaseUserSyncLock(testUserID)
}
