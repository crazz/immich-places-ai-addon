package main

import (
	"context"
	"net/http"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorSendRechecksStandardFieldsImmediatelyBeforeMutation(t *testing.T) {
	f := mirrorWriteFixture(t)
	mirrorMutationServer(t, f)
	ctx := context.Background()
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "freshness"})
	if err != nil {
		t.Fatal(err)
	}
	if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	reads := 0
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == "GET" && r.URL.Path == "/api/assets/"+f.draft.AssetID+"/metadata" {
			reads++
			if reads == 1 {
				f.mu.Lock()
				f.meta["exifInfo"].(map[string]any)["description"] = "concurrent user edit"
				f.mu.Unlock()
			}
		}
		return false
	}
	if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if f.mirrorSends != 0 || !saved.Targets[0].Verified || saved.Status == "succeeded" {
		t.Fatal("metadata sent after prerequisite changed", f.mirrorSends, saved.Status)
	}
}
