package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAIProviderHTTPFailuresAreSafeAndLeaveNoWrites(t *testing.T) {
	for _, scenario := range []string{"content type", "duplicate origin", "invalid session", "expired session", "create revision", "missing edit revision", "storage list", "storage save", "storage auth"} {
		t.Run(scenario, func(t *testing.T) {
			db, handler := aiTestHandler(t, true)
			input := aiProviderRequest{Input: providerInput()}
			method, target, expected, code := "POST", "/ai/providers", 400, "INVALID_REQUEST"
			if scenario == "create revision" {
				input.ExpectedRevision = 1
			}
			if scenario == "missing edit revision" {
				method, target, code = "PUT", "/ai/providers/profile", "INVALID_PROVIDER"
			}
			if scenario == "expired session" {
				if _, err := db.db.Exec("UPDATE sessions SET expiresAt='2000-01-01T00:00:00Z'"); err != nil {
					t.Fatal(err)
				}
			}
			if strings.HasPrefix(scenario, "storage") {
				if _, err := db.db.Exec("DROP TABLE ai_provider_versions"); err != nil {
					t.Fatal(err)
				}
				expected, code = 500, "STORAGE_ERROR"
				if scenario == "storage list" {
					method = "GET"
				}
				if scenario == "storage auth" {
					if err := db.db.Close(); err != nil {
						t.Fatal(err)
					}
				}
			}
			body, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(method, target, strings.NewReader(string(body)))
			req.Header.Set("Origin", aiTestOrigin)
			req.Header.Set("Content-Type", "application/json")
			cookie := "ai-session"
			switch scenario {
			case "content type":
				req.Header.Set("Content-Type", "text/plain")
				expected = 415
			case "duplicate origin":
				req.Header.Add("Origin", aiTestOrigin)
				expected, code = 403, "ORIGIN_REJECTED"
			case "invalid session":
				cookie = "invalid"
				expected, code = 401, "UNAUTHENTICATED"
			case "expired session":
				expected, code = 401, "UNAUTHENTICATED"
			}
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: cookie})
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			var failure struct {
				Code, Message, RequestID string
				Retryable                bool
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &failure); err != nil || rec.Code != expected || failure.Code != code || failure.RequestID == "" || failure.Retryable != (expected == 500) {
				t.Fatalf("unexpected failure: %d %s, %v", rec.Code, rec.Body.String(), err)
			}
			if strings.Contains(rec.Body.String(), *input.Secret) || strings.Contains(rec.Body.String(), "ai_provider_versions") {
				t.Fatal("response exposed a secret or internal storage details")
			}
			if scenario != "storage auth" {
				var count int
				if err := db.db.QueryRowContext(context.Background(), "SELECT count(*) FROM ai_provider_profiles").Scan(&count); err != nil || count != 0 {
					t.Fatalf("rejected mutation persisted: count=%d, %v", count, err)
				}
			}
		})
	}
}
