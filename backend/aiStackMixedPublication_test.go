package main

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
)

func TestAIStackMixedSuccessConflictAndMissingCatalogRemainIndependentInHistory(t *testing.T) {
	f := stackWriteFixture(t)
	op := approveStackFixture(t, f, f.members)
	conflict, missing := op.Targets[1].AssetID, op.Targets[2].AssetID
	f.metadata[conflict]["exifInfo"].(map[string]any)["latitude"] = 25
	m := stackMutationServer(t, f)
	m.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == "PATCH" && r.URL.Path == "/api/assets/"+missing {
			selectionSQL(t, f.f.db, `DELETE FROM assets WHERE userID=? AND immichID=?`, testUserID, missing)
		}
		return false
	}
	ctx := context.Background()
	for range 3 {
		if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
			t.Fatal(err)
		}
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || saved.Verified || saved.Refreshed || saved.Status != "verifying" || saved.Digest != op.Digest {
		t.Fatal("mixed publication claimed complete success", err)
	}
	if saved.Targets[0].Status != "succeeded" || !saved.Targets[0].Verified || !saved.Targets[0].Refreshed || m.sent(saved.Targets[0].AssetID) != 1 {
		t.Fatal("verified target lost independent success")
	}
	if saved.Targets[1].Status != "conflict" || saved.Targets[1].Verified || saved.Targets[1].Refreshed || m.sent(conflict) != 0 {
		t.Fatal("changed target was overwritten or refreshed")
	}
	if saved.Targets[2].Status != "verifying" || !saved.Targets[2].Verified || saved.Targets[2].Refreshed || saved.Targets[2].Code != "LOCAL_REFRESH_PENDING" || m.sent(missing) != 1 {
		t.Fatal("missing row hid verified remote state")
	}
	for i, target := range saved.Targets {
		var lat, lon sql.NullFloat64
		err := f.f.db.db.QueryRow(`SELECT latitude,longitude FROM assets WHERE userID=? AND immichID=?`, testUserID, target.AssetID).Scan(&lat, &lon)
		if i == 2 {
			if err != sql.ErrNoRows {
				t.Fatal("missing row was fabricated", err)
			}
		} else if err != nil || i == 0 && (!lat.Valid || !lon.Valid || lat.Float64 != 0 || lon.Float64 != 12) || i == 1 && (lat.Valid || lon.Valid) {
			t.Fatal("catalog refresh ignored target verification", err)
		}
	}
	f.writer.enabled = func() bool { return false }
	page, err := f.writer.history(ctx, testUserID, f.draft.ID, "")
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != op.ID || page.Items[0].Status != saved.Status {
		t.Fatal("mixed target history lost on disabled reopen", err)
	}
	retained, err := f.writer.get(ctx, testUserID, page.Items[0].ID, false)
	if err != nil || len(retained.Targets) != 3 || len(retained.Targets[0].Events) == 0 || retained.Targets[2].Fields[0].Status != "verified" {
		t.Fatal("retained audit hid independent target evidence", err)
	}
}
