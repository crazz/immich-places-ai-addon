package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorStandardStepVerifiesBeforeMirrorAndRetainsGuard(t *testing.T) {
	f := mirrorWriteFixture(t)
	mirrorMutationServer(t, f)
	ctx := context.Background()
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method != http.MethodPatch {
			return false
		}
		f.mu.Lock()
		defer f.mu.Unlock()
		var fields map[string]any
		if json.NewDecoder(r.Body).Decode(&fields) != nil || len(fields) != 3 {
			t.Error("standard fields expanded")
		}
		f.meta["exifInfo"] = fields
		f.standardSends++
		w.WriteHeader(200)
		return true
	}
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "standard-first"})
	if err != nil {
		t.Fatal(err)
	}
	if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	op, err = f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || !op.Targets[0].Verified || !op.Targets[0].Refreshed || op.Mirror.Status != "blocked" || op.Verified || op.Status == "succeeded" {
		t.Fatalf("standard/mirror state not independent: error=%v standard=%s/%s verified=%t refreshed=%t mirror=%s aggregate=%s verified=%t", err, op.Targets[0].Status, op.Targets[0].Code, op.Targets[0].Verified, op.Targets[0].Refreshed, op.Mirror.Status, op.Status, op.Verified)
	}
	if f.standardSends != 1 || f.mirrorSends != 0 {
		t.Fatal("mirror sent before its own reservation")
	}
	var guards int
	if err = f.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards WHERE assetID=?`, f.draft.AssetID).Scan(&guards); err != nil || guards != 1 {
		t.Fatal("optional step lost target exclusion", err)
	}
}
