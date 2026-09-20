package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAIProductionHTTPRejectsUnprotectedAndAmbiguousRequests(t *testing.T) {
	f, p, req := productionFixture(t)
	productionSession(t, f)
	h := newAIJobHandler(p, aiTestOrigin)
	data, _ := json.Marshal(req)
	valid := string(data)
	cases := []struct {
		name, body, origin string
		auth               bool
		status             int
	}{
		{"session", valid, aiTestOrigin, false, 401},
		{"origin", valid, "", true, 403},
		{"foreign-origin", valid, "https://foreign.example", true, 403},
		{"owner", strings.TrimSuffix(valid, "}") + `,"owner":"other"}`, aiTestOrigin, true, 400},
		{"membership", strings.TrimSuffix(valid, "}") + `,"assetIds":[]}`, aiTestOrigin, true, 400},
		{"trailing", valid + " {}", aiTestOrigin, true, 400},
		{"oversized", valid + strings.Repeat(" ", 32<<10), aiTestOrigin, true, 400},
		{"duplicate", strings.TrimSuffix(valid, "}") + `,"idempotencyKey":"other"}`, aiTestOrigin, true, 400},
		{"case-duplicate", strings.TrimSuffix(valid, "}") + `,"IDEMPOTENCYKEY":"other"}`, aiTestOrigin, true, 400},
		{"null", "null", aiTestOrigin, true, 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := aiRequest(h, "POST", "/ai/jobs", tc.body, tc.origin, tc.auth)
			if rec.Code != tc.status || rec.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("%d %s", rec.Code, rec.Body.String())
			}
		})
	}
	for _, id := range []string{"foreign-private-id", "absent-id"} {
		rec := aiRequest(h, "GET", "/ai/jobs/"+id, "", "", true)
		if rec.Code != 404 || strings.Contains(rec.Body.String(), id) {
			t.Fatal("existence leak", rec.Body.String())
		}
	}
	var count int
	if err := f.image.db.db.QueryRow("SELECT count(*) FROM ai_jobs").Scan(&count); err != nil || count != 0 {
		t.Fatal("rejected HTTP input admitted", count, err)
	}
}
