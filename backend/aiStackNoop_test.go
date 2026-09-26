package main

import (
	"context"
	"testing"
)

func TestAIStackMatchingSiblingIsVerifiedWithoutAMutationAttempt(t *testing.T) {
	f := stackWriteFixture(t)
	f.metadata[selectionB]["exifInfo"] = map[string]any{"latitude": 0, "longitude": 12, "description": "sibling private text"}
	op := approveStackFixture(t, f, f.members)
	m := stackMutationServer(t, f)
	ctx := context.Background()
	for range 4 {
		if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
			t.Fatal(err)
		}
	}
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || saved.Status != "succeeded" || !saved.Verified || !saved.Refreshed || saved.Noop {
		t.Fatal("no-op contaminated aggregate outcomes", err)
	}
	for _, target := range saved.Targets {
		wantAttempts := 1
		if target.AssetID == selectionB {
			wantAttempts = 0
			if !target.Noop || !target.Settled || !target.Verified || !target.Refreshed || target.Fields[0].Status != "verified" {
				t.Fatal("no-op not independently verified")
			}
		}
		if target.Attempts != wantAttempts || m.sent(target.AssetID) != wantAttempts {
			t.Fatal("already matching sibling reserved or sent a mutation")
		}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.metadata[selectionB]["exifInfo"].(map[string]any)["description"] != "sibling private text" {
		t.Fatal("GPS no-op changed sibling text")
	}
}
