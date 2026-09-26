package main

import (
	"context"
	"net/http"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorReadbackRejectsSourceChangedDuringNamespaceRead(t *testing.T) {
	f := mirrorWriteFixture(t)
	mirrorMutationServer(t, f)
	ctx := context.Background()
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "source-readback"})
	if err != nil {
		t.Fatal(err)
	}
	if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == "GET" && r.URL.Path == "/api/assets/"+f.draft.AssetID+"/metadata" {
			f.mu.Lock()
			if f.mirrorSends == 1 {
				f.meta["checksum"] = "changed-after-first-image-read"
			}
			f.mu.Unlock()
		}
		return false
	}
	_ = f.writer.runOne(ctx, testUserID, op.ID)
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Mirror.Verified || saved.Status == "succeeded" || f.mirrorSends != 1 {
		t.Fatal("namespace from a changed source was verified", saved.Mirror, f.mirrorSends)
	}
	var published int
	if err = f.f.db.db.QueryRow(`SELECT count(*) FROM ai_mirror_records WHERE lastOperationID IS NOT NULL`).Scan(&published); err != nil || published != 0 {
		t.Fatal("inconsistent read granted ownership", err)
	}
}
