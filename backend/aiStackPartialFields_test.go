package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
)

func TestAIStackPartialPrimaryFieldsRemainIndependentAndNeverReplay(t *testing.T) {
	for _, applied := range []string{"gps", "description"} {
		t.Run(applied, func(t *testing.T) {
			f := stackWriteFixture(t)
			op := approveStackFixture(t, f, f.members)
			m := stackMutationServer(t, f)
			m.handle = func(w http.ResponseWriter, r *http.Request) bool {
				if r.Method != "PATCH" || r.URL.Path != "/api/assets/"+op.Plan.TargetID {
					return false
				}
				var payload map[string]any
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Error(err)
				}
				f.mu.Lock()
				exif := f.metadata[op.Plan.TargetID]["exifInfo"].(map[string]any)
				if applied == "gps" {
					exif["latitude"], exif["longitude"] = payload["latitude"], payload["longitude"]
				} else {
					exif["description"] = payload["description"]
				}
				f.mu.Unlock()
				w.WriteHeader(200)
				return true
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
			if err != nil || saved.Status != "partial" || saved.Verified {
				t.Fatal("partial result reported full success", err)
			}
			for _, target := range saved.Targets {
				if target.AssetID == op.Plan.TargetID {
					if target.Status != "partial" || target.Verified || target.Refreshed != (applied == "gps") {
						t.Fatal("primary field outcomes collapsed", target.Status)
					}
					for _, field := range target.Fields {
						want := "baseline"
						if field.Field == applied {
							want = "verified"
						}
						if field.Status != want || field.WasVerified != (field.Field == applied) {
							t.Fatal("field evidence changed on reopen", field.Field, field.Status)
						}
					}
					if _, err := f.writer.retryStackTarget(ctx, testUserID, op.ID, target.AssetID, target.Generation); err == nil {
						t.Fatal("partially successful payload became retryable")
					}
				} else if target.Status != "succeeded" || !target.Verified || !target.Refreshed {
					t.Fatal("primary partial outcome contaminated sibling", target.Status)
				}
				if m.sent(target.AssetID) != 1 {
					t.Fatal("target implicitly resent")
				}
			}
			var lat, lon sql.NullFloat64
			if err := f.f.db.db.QueryRow(`SELECT latitude,longitude FROM assets WHERE userID=? AND immichID=?`, testUserID, op.Plan.TargetID).Scan(&lat, &lon); err != nil {
				t.Fatal(err)
			}
			if applied == "gps" && (!lat.Valid || !lon.Valid || lat.Float64 != 0 || lon.Float64 != 12) || applied == "description" && (lat.Valid || lon.Valid) {
				t.Fatal("catalog did not follow GPS-specific verification")
			}
		})
	}
}
