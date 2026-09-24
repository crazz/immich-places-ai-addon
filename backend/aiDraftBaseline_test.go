package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func baselineRequest(h http.Handler, id string, revision int, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", "/ai/drafts/"+id+"/baseline", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", aiTestOrigin)
	req.Header.Set("If-Match", strconv.Quote(strconv.Itoa(revision)))
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "ai-session"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
func TestAIDraftBaselineRetainsExactNullableGPSAndOriginalProvenance(t *testing.T) {
	for _, gps := range []map[string]any{{"latitude": nil, "longitude": nil}, {"latitude": 0, "longitude": nil}, {"latitude": 0, "longitude": 0}, {"latitude": 30, "longitude": 40}} {
		f, id := draftFixture(t)
		value := acceptedDraft(t, f, id)
		image := newAIImageFixture(t)
		seedAsset(t, f.db, selectionA, nil, nil, "2026-09-20")
		key := "private-immich-image-key"
		if err := f.db.updateImmichAPIKey(context.Background(), testUserID, &key); err != nil {
			t.Fatal(err)
		}
		image.handle = func(w http.ResponseWriter, r *http.Request) bool {
			if r.URL.Path != "/api/assets/"+selectionA {
				return false
			}
			meta := image.metadata()
			meta["exifInfo"] = gps
			if err := json.NewEncoder(w).Encode(meta); err != nil {
				t.Error(err)
			}
			return true
		}
		h := newAIResultHandler(&aiResultStore{jobs: f.store, origin: aiTestOrigin}, image.service)
		prepared := baselineRequest(h, value.ID, 1, `{}`)
		if prepared.Code != 200 {
			t.Fatalf("prepare %d %s", prepared.Code, prepared.Body.String())
		}
		var observation struct {
			ID       string `json:"id"`
			Baseline struct {
				Latitude      *float64 `json:"latitude"`
				Longitude     *float64 `json:"longitude"`
				ImageIdentity string   `json:"imageIdentity"`
			} `json:"baseline"`
		}
		if json.Unmarshal(prepared.Body.Bytes(), &observation) != nil || observation.ID == "" || observation.Baseline.ImageIdentity == "" {
			t.Fatal("missing observation")
		}
		acknowledged := baselineRequest(h, value.ID, 1, `{"observationId":"`+observation.ID+`"}`)
		if acknowledged.Code != 200 {
			t.Fatal("acknowledge", acknowledged.Code, acknowledged.Body.String())
		}
		var saved map[string]any
		if json.Unmarshal(acknowledged.Body.Bytes(), &saved) != nil || saved["revision"] != float64(2) || saved["originalSourceDigest"] != value.OriginalSourceDigest {
			t.Fatal("original or revision changed", saved)
		}
		baseline := saved["baseline"].(map[string]any)
		for _, field := range []string{"latitude", "longitude"} {
			actual, _ := json.Marshal(baseline[field])
			expected, _ := json.Marshal(gps[field])
			if string(actual) != string(expected) {
				t.Fatal("GPS semantics lost", field, string(actual), string(expected))
			}
		}
		if _, bad := image.counts(); bad != 0 {
			t.Fatal("upstream mutation or wrong authority", bad)
		}
	}
}
