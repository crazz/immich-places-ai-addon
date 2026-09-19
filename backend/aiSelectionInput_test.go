package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAISelectionRejectsAmbiguousJSON(t *testing.T) {
	db, _ := aiTestHandler(t, true)
	h := newAISelectionHandler(db, &Config{AIEnabled: true, ImmichURL: "https://immich.example/api", AIPublicOrigin: aiTestOrigin})
	for _, tc := range []struct {
		body, content string
		status        int
	}{
		{"null", "application/json", 400},
		{selectionBody + ` {}`, "application/json", 400},
		{strings.TrimSuffix(selectionBody, "}") + `,"mode":"explicit"}`, "application/json", 400},
		{strings.Replace(selectionBody, `"view":"all"`, `"view":"all","page":2`, 1), "application/json", 400},
		{strings.Replace(selectionBody, `"view":"all"`, `"view":"all","albumID":""`, 1), "application/json", 400},
		{strings.Replace(selectionBody, `"view":"all"`, `"view":"all","gpsFilter":null`, 1), "application/json", 400},
		{selectionBody + strings.Repeat(" ", 1<<20), "application/json", 413},
		{selectionBody, "text/plain", 415},
	} {
		req := httptest.NewRequest("POST", "/ai/selection-preview", strings.NewReader(tc.body))
		req.Header.Set("Origin", aiTestOrigin)
		req.Header.Set("Content-Type", tc.content)
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "ai-session"})
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != tc.status {
			t.Fatalf("status %d want %d for %.100s: %s", rec.Code, tc.status, tc.body, rec.Body.String())
		}
	}
	var n int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&n); err != nil || n != 0 {
		t.Fatalf("persisted rejected JSON: %d %v", n, err)
	}
}

func TestAISelectionRejectsDuplicateOrigin(t *testing.T) {
	db, _ := aiTestHandler(t, true)
	h := newAISelectionHandler(db, &Config{AIEnabled: true, ImmichURL: "https://immich.example/api", AIPublicOrigin: aiTestOrigin})
	req := httptest.NewRequest("POST", "/ai/selection-preview", strings.NewReader(selectionBody))
	req.Header.Add("Origin", aiTestOrigin)
	req.Header.Add("Origin", aiTestOrigin)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "ai-session"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var count int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 403 || count != 0 || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("duplicate origin accepted: %d count=%d", rec.Code, count)
	}
}
