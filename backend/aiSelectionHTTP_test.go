package main

import (
	"encoding/json"
	"immich-places-backend/internal/ai/selection"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestAISelectionRequiresSession(t *testing.T) {
	db := newTestDB(t)
	handler := newAISelectionHandler(db, &Config{AIEnabled: true, ImmichURL: "https://immich.example/api", AIPublicOrigin: aiTestOrigin})
	for _, path := range []string{"/ai/selection-preview", "/ai/selections/private-id"} {
		rec := aiRequest(handler, http.MethodPost, path, `{}`, aiTestOrigin, false)
		if rec.Code != http.StatusUnauthorized || rec.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("unprotected selection: %d %s", rec.Code, rec.Body.String())
		}
	}
}

func TestAISelectionAuthorityGuards(t *testing.T) {
	db, _ := aiTestHandler(t, true)
	for _, tc := range []struct {
		enabled bool
		origin  string
		status  int
	}{
		{false, aiTestOrigin, 503}, {true, "", 403}, {true, "https://foreign.example", 403},
	} {
		h := newAISelectionHandler(db, &Config{AIEnabled: tc.enabled, ImmichURL: "https://immich.example/api", AIPublicOrigin: aiTestOrigin})
		rec := aiRequest(h, http.MethodPost, "/ai/selection-preview", `{}`, tc.origin, true)
		if rec.Code != tc.status {
			t.Fatalf("got %d want %d: %s", rec.Code, tc.status, rec.Body.String())
		}
	}
	h := newAISelectionHandler(db, &Config{})
	if rec := aiRequest(h, http.MethodGet, "/ai/selections/id", "", "", true); rec.Code != 503 {
		t.Fatalf("disabled read: %d", rec.Code)
	}
}

const selectionA = "aaaaaaaa-0000-4000-8000-000000000001"
const selectionB = "bbbbbbbb-0000-4000-8000-000000000002"
const selectionBody = `{"mode":"explicit","assetIDs":["` + selectionA + `","` + selectionB + `","` + selectionA + `"],"scope":{"view":"all"}}`

func TestAISelectionOwnerRoundTrip(t *testing.T) {
	db, _ := aiTestHandler(t, true)
	seedAsset(t, db, selectionA, nil, nil, "2026-09-19T23:00:00-12:00")
	seedAsset(t, db, selectionB, nil, nil, "2026-09-18T23:00:00")
	h := newAISelectionHandler(db, &Config{AIEnabled: true, ImmichURL: "https://immich.example/api", AIPublicOrigin: aiTestOrigin})
	rec := aiRequest(h, "POST", "/ai/selection-preview", selectionBody, aiTestOrigin, true)
	if rec.Code != 200 {
		t.Fatalf("preview: %d %s", rec.Code, rec.Body.String())
	}
	var result struct {
		SnapshotID                                                                string
		AssetIDs                                                                  []string
		RequestedCount, UniqueCount, DuplicateCount, EligibleCount, ExcludedCount int
		ExpiresAt                                                                 string
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.SnapshotID == "" || result.ExpiresAt == "" || strings.Join(result.AssetIDs, ",") != selectionA+","+selectionB || result.RequestedCount != 3 || result.UniqueCount != 2 || result.DuplicateCount != 1 || result.EligibleCount != 2 || result.ExcludedCount != 0 {
		t.Fatalf("manifest: %s", rec.Body.String())
	}
	read := aiRequest(h, "GET", "/ai/selections/"+result.SnapshotID, "", "", true)
	if read.Code != 200 || read.Body.String() != rec.Body.String() || read.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("read: %d %s", read.Code, read.Body.String())
	}
	var count int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_items").Scan(&count); err != nil || count != 2 {
		t.Fatalf("persistence count=%d err=%v", count, err)
	}
}

func TestAISelectionUsesConfiguredLimits(t *testing.T) {
	db, _ := aiTestHandler(t, true)
	seedAsset(t, db, selectionA, nil, nil, "2026-09-19")
	seedAsset(t, db, selectionB, nil, nil, "2026-09-19")
	cfg := &Config{AIEnabled: true, ImmichURL: "https://immich.example/api", AIPublicOrigin: aiTestOrigin, AISelectionMaxAssets: 1, AISelectionTTLSeconds: 60, AIInstanceEpoch: "2"}
	h := newAISelectionHandler(db, cfg)
	rec := aiRequest(h, "POST", "/ai/selection-preview", selectionBody, aiTestOrigin, true)
	if rec.Code != 413 {
		t.Fatalf("configured limit ignored: %d", rec.Code)
	}
	body := strings.ReplaceAll(selectionBody, selectionB, selectionA)
	rec = aiRequest(h, "POST", "/ai/selection-preview", body, aiTestOrigin, true)
	var result selection.Manifest
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 || result.ExpiresAt.Sub(result.CreatedAt) != time.Minute {
		t.Fatalf("configured ttl ignored: %s", rec.Body.String())
	}
	cfg.AIInstanceEpoch = "3"
	h = newAISelectionHandler(db, cfg)
	rec = aiRequest(h, "GET", "/ai/selections/"+*result.SnapshotID, "", "", true)
	if rec.Code != 404 {
		t.Fatalf("configured epoch ignored: %d", rec.Code)
	}
}
