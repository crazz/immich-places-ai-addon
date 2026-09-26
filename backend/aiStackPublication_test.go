package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestAIStackMissingCatalogTargetRetainsVerificationAndRecoversReadOnly(t *testing.T) {
	f := stackWriteFixture(t)
	op := approveStackFixture(t, f, f.members)
	m := stackMutationServer(t, f)
	asset := op.Targets[0].AssetID
	m.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == "PATCH" && r.URL.Path == "/api/assets/"+asset {
			selectionSQL(t, f.f.db, `DELETE FROM assets WHERE userID=? AND immichID=?`, testUserID, asset)
		}
		return false
	}
	ctx := context.Background()
	if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || !saved.Targets[0].Verified || saved.Targets[0].Refreshed || saved.Targets[0].Status != "verifying" || saved.Targets[0].Code != "LOCAL_REFRESH_PENDING" {
		t.Fatal("upstream verification and local refresh were conflated", err)
	}
	var rows int
	if err := f.f.db.db.QueryRow(`SELECT count(*) FROM assets WHERE userID=? AND immichID=?`, testUserID, asset).Scan(&rows); err != nil || rows != 0 {
		t.Fatal("writer invented missing catalog row", err)
	}
	seedAsset(t, f.f.db, asset, nil, nil, "2026-09-20")
	f.f.now = f.f.now.Add(2 * time.Second)
	if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	saved, err = f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || saved.Targets[0].Status != "succeeded" || !saved.Targets[0].Refreshed || m.sent(asset) != 1 {
		t.Fatal("local repair resent or lost verification", err)
	}
	for _, target := range saved.Targets[1:] {
		if target.Attempts != 0 || target.Status != "queued" || m.sent(target.AssetID) != 0 {
			t.Fatal("one target repair changed another target")
		}
	}
}
