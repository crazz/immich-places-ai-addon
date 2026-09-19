package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const aiTestOrigin = "https://places.example"

func aiTestHandler(t *testing.T, enabled bool) (*Database, http.Handler) {
	t.Helper()
	db := newTestDB(t)
	hash := sha256.Sum256([]byte("ai-session"))
	if err := db.createSession(context.Background(), hex.EncodeToString(hash[:]), testUserID, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	return db, newAIProviderHandler(db, &Config{AIEnabled: enabled, AIPublicOrigin: aiTestOrigin}, nil)
}

func aiRequest(handler http.Handler, method, path, body, origin string, authenticated bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", origin)
	if authenticated {
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "ai-session"})
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestAIProviderHTTPProtection(t *testing.T) {
	valid := `{"name":"Private","baseURL":"https://provider.example/v1","model":"manual-model","enabled":true,"secret":"private-secret"}`
	for _, tc := range []struct {
		name, origin, body, code string
		auth, enabled            bool
		status                   int
	}{
		{"create", aiTestOrigin, valid, "", true, true, 201},
		{"no session", aiTestOrigin, valid, "UNAUTHENTICATED", false, true, 401},
		{"disabled", aiTestOrigin, valid, "AI_DISABLED", true, false, 503},
		{"missing origin", "", valid, "ORIGIN_REJECTED", true, true, 403},
		{"null origin", "null", valid, "ORIGIN_REJECTED", true, true, 403},
		{"foreign origin", "https://attacker.example", valid, "ORIGIN_REJECTED", true, true, 403},
		{"bad JSON", aiTestOrigin, "{", "INVALID_REQUEST", true, true, 400},
		{"null JSON", aiTestOrigin, "null", "INVALID_REQUEST", true, true, 400},
		{"trailing JSON", aiTestOrigin, valid + " {}", "INVALID_REQUEST", true, true, 400},
		{"unknown option", aiTestOrigin, strings.TrimSuffix(valid, "}") + `,"temperature":0.5}`, "INVALID_REQUEST", true, true, 400},
		{"invalid config", aiTestOrigin, `{"name":"","baseURL":"https://provider.example","model":"vision"}`, "INVALID_PROVIDER", true, true, 400},
		{"oversized", aiTestOrigin, valid + strings.Repeat(" ", 16384), "INVALID_REQUEST", true, true, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, handler := aiTestHandler(t, tc.enabled)
			rec := aiRequest(handler, "POST", "/ai/providers", tc.body, tc.origin, tc.auth)
			if rec.Code != tc.status {
				t.Fatalf("status = %d, body = %s; want %d", rec.Code, rec.Body.String(), tc.status)
			}
			if strings.Contains(rec.Body.String(), "private-secret") {
				t.Fatal("secret leaked in response")
			}
			var profiles int
			if err := db.db.QueryRow("SELECT count(*) FROM ai_provider_profiles").Scan(&profiles); err != nil {
				t.Fatal(err)
			}
			if tc.status == 201 {
				if profiles != 1 {
					t.Fatal("successful save was not persisted")
				}
				return
			}
			var failure struct {
				Code, Message, RequestID string
				Retryable                bool
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &failure); err != nil || failure.Code != tc.code || failure.Message == "" || failure.RequestID == "" || profiles != 0 {
				t.Fatalf("unsafe error or persisted rejected request: %s, profiles = %d, error = %v", rec.Body.String(), profiles, err)
			}
		})
	}
}
