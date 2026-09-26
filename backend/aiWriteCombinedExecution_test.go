package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStandardCombinedSocketWritesExactFieldsAndReconcilesEachOutcome(t *testing.T) {
	for _, scenario := range []struct {
		name                           string
		gps, description, loseResponse bool
		status                         string
	}{
		{"both", true, true, false, "succeeded"},
		{"gps-only", true, false, false, "partial"},
		{"description-only", false, true, false, "partial"},
		{"lost-response", true, true, true, "succeeded"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			w := standardWriteFixture(t, []string{"gps", "description"})
			ctx := context.Background()
			op, err := w.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: scenario.name})
			if err != nil {
				t.Fatal(err)
			}
			var mu sync.Mutex
			sends := 0
			server := httptest.NewServer(http.HandlerFunc(func(out http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				if r.URL.Path != "/api/assets/"+op.Plan.TargetID || r.Header.Get("x-api-key") != "private-immich-image-key" {
					t.Error("wrong mutation authority")
					out.WriteHeader(403)
					return
				}
				if r.Method == "PATCH" {
					sends++
					var payload map[string]json.RawMessage
					if json.NewDecoder(r.Body).Decode(&payload) != nil || len(payload) != 3 {
						t.Error("combined request fields differ from approval")
						out.WriteHeader(400)
						return
					}
					var latitude, longitude float64
					var description string
					if json.Unmarshal(payload["latitude"], &latitude) != nil || json.Unmarshal(payload["longitude"], &longitude) != nil || json.Unmarshal(payload["description"], &description) != nil || latitude != op.Plan.Intended.Latitude || longitude != op.Plan.Intended.Longitude || description != op.Plan.Description.Intended {
						t.Error("combined payload changed approved values")
						out.WriteHeader(400)
						return
					}
					exif := map[string]any{"latitude": nil, "longitude": nil, "description": op.Plan.Description.Before.Value}
					if scenario.gps {
						exif["latitude"], exif["longitude"] = latitude, longitude
					}
					if scenario.description {
						exif["description"] = description
					}
					w.meta["exifInfo"] = exif
					if scenario.loseResponse {
						conn, _, err := out.(http.Hijacker).Hijack()
						if err != nil {
							t.Error(err)
							return
						}
						conn.Close()
						return
					}
				} else if r.Method != "GET" {
					t.Error("unexpected request method")
				}
				if err := json.NewEncoder(out).Encode(w.meta); err != nil {
					t.Error(err)
				}
			}))
			t.Cleanup(server.Close)
			w.image.service.endpoint = server.URL
			if err := w.writer.runOne(ctx, testUserID, op.ID); err != nil {
				t.Fatal(err)
			}
			w.f.reopen(t)
			w.writer.drafts.results.jobs = w.f.store
			current, err := w.writer.get(ctx, testUserID, op.ID, false)
			if err != nil || current.Status != scenario.status || current.Attempts != 1 || current.Verified != (scenario.gps && scenario.description) || current.Refreshed != scenario.gps || len(current.Fields) != 2 {
				t.Fatal("wrong combined outcome", current.Status, err)
			}
			for i, verified := range []bool{scenario.gps, scenario.description} {
				if (current.Fields[i].Status == "verified") != verified || current.Fields[i].WasVerified != verified {
					t.Fatal("field outcome lost", current.Fields)
				}
			}
			var latitude, longitude sql.NullFloat64
			if err := w.f.db.db.QueryRow(`SELECT latitude,longitude FROM assets WHERE userID=? AND immichID=?`, testUserID, op.Plan.TargetID).Scan(&latitude, &longitude); err != nil || latitude.Valid != scenario.gps || longitude.Valid != scenario.gps {
				t.Fatal("GPS refresh followed wrong field", err)
			}
			if _, err := w.writer.retry(ctx, testUserID, op.ID, current.Generation); err == nil {
				t.Fatal("combined successful field could replay")
			}
			mu.Lock()
			count := sends
			mu.Unlock()
			if count != 1 {
				t.Fatal("hidden mutation retry", count)
			}
		})
	}
}
