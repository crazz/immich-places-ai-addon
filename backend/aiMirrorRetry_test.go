package main

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorRetryOnlyResendsMetadataAndAcceptedReplayIsLocal(t *testing.T) {
	f := mirrorWriteFixture(t)
	mirrorMutationServer(t, f)
	ctx := context.Background()
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == "PUT" {
			f.mu.Lock()
			f.mirrorSends++
			f.mu.Unlock()
			w.WriteHeader(400)
			return true
		}
		return false
	}
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "retry-mirror"})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
			t.Fatal(err)
		}
	}
	op, err = f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || op.Mirror.Status != "retryable" || op.Status != "partial" || !op.Targets[0].Verified {
		t.Fatal("failed optional step hid standard success", err)
	}
	generation := op.Mirror.Generation
	accepted, err := f.writer.retryMirror(ctx, testUserID, op.ID, f.draft.AssetID, generation)
	if err != nil || accepted.Mirror.Status != "queued" || accepted.Mirror.Generation != generation+1 || accepted.Mirror.Attempts != 1 {
		t.Fatal("exact retry not accepted", err)
	}
	f.handle = nil
	if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || !saved.Verified || saved.Mirror.Attempts != 2 {
		t.Fatal("retry did not finish metadata", err)
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	f.writer.enabled = func() bool { return false }
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		t.Error("accepted replay performed I/O")
		w.WriteHeader(503)
		return true
	}
	replay, err := f.writer.retryMirror(ctx, testUserID, op.ID, f.draft.AssetID, generation)
	if err != nil || !reflect.DeepEqual(saved, replay) {
		t.Fatal("accepted-generation replay changed after success/reopen/disable", err)
	}
	if f.standardSends != 1 || f.mirrorSends != 2 {
		t.Fatal("metadata repair resent standard work")
	}
}
