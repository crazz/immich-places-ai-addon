package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestExplicitlyTestExistingProxyWithSelectedModel(t *testing.T) {
	var received struct {
		hits                                        atomic.Int32
		immichHits                                  atomic.Int32
		modelsHits                                  atomic.Int32
		method, path, auth, contentType, accept, ua string
		body                                        []byte
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received.hits.Add(1)
		if strings.Contains(r.URL.Path, "/models") {
			received.modelsHits.Add(1)
		}
		received.method = r.Method
		received.path = r.URL.Path
		received.auth = r.Header.Get("Authorization")
		received.contentType = r.Header.Get("Content-Type")
		received.accept = r.Header.Get("Accept")
		received.ua = r.Header.Get("User-Agent")
		received.body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		bodyText := string(received.body)
		content := "color: blue\nshape: circle"
		if strings.Contains(bodyText, `"json_object"`) || strings.Contains(bodyText, `"json_schema"`) {
			content = `{"color":"blue","shape":"circle"}`
		}
		contentJSON, _ := json.Marshal(content)
		_, _ = w.Write([]byte(`{"id":"cmpl-1","model":"gpt-5.6-sol","choices":[{"message":{"role":"assistant","content":` + string(contentJSON) + `},"finish_reason":"stop"}]}`))
	}))
	proxy.Listener = listener
	proxy.Start()
	t.Cleanup(proxy.Close)

	immich := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		received.immichHits.Add(1)
	}))
	t.Cleanup(immich.Close)

	base, err := url.Parse(proxy.URL)
	if err != nil {
		t.Fatal(err)
	}
	baseURL := "http://codex-proxy:" + base.Port() + "/v1"
	policyJSON := `[{"baseURL":"` + baseURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`
	policy, err := providers.ParseEgressPolicy(policyJSON)
	if err != nil {
		t.Fatal(err)
	}

	transport := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})
	_, handler := aiCapabilityTestHandler(t, true, policy, transport)

	secret := "proxy-secret"
	created := aiRequest(handler, "POST", "/ai/providers", mustJSON(t, providers.Input{
		Config:  providers.Config{Name: "NAS Proxy", BaseURL: baseURL, Model: "gpt-5.6-sol"},
		Enabled: true,
		Secret:  &secret,
	}), aiTestOrigin, true)
	var profile aiProviderProfile
	if err := json.Unmarshal(created.Body.Bytes(), &profile); err != nil || created.Code != 201 {
		t.Fatalf("create profile = %s err=%v", created.Body.String(), err)
	}

	rec := aiRequest(handler, "POST", "/ai/providers/"+profile.ID+"/test", mustJSON(t, map[string]any{
		"expectedRevision": profile.Revision,
	}), aiTestOrigin, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("capability test status = %d body = %s", rec.Code, rec.Body.String())
	}

	if received.hits.Load() != 3 {
		t.Fatalf("proxy hits = %d, want exactly three synthetic test requests", received.hits.Load())
	}
	if received.modelsHits.Load() != 0 {
		t.Fatal("model-list request occurred")
	}
	if received.immichHits.Load() != 0 {
		t.Fatal("Immich request occurred")
	}
	if received.method != http.MethodPost || received.path != "/v1/chat/completions" {
		t.Fatalf("unexpected proxy request: method=%s path=%s", received.method, received.path)
	}
	if received.auth != "Bearer proxy-secret" {
		t.Fatalf("unexpected authorization: %q", received.auth)
	}

	var payload map[string]any
	if err := json.Unmarshal(received.body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["model"] != "gpt-5.6-sol" {
		t.Fatalf("selected model not used: %v", payload["model"])
	}
	if payload["stream"] != false {
		t.Fatalf("stream must be false: %v", payload["stream"])
	}
	bodyText := string(received.body)
	for _, forbidden := range []string{"immich", "photo", "library", immich.URL, "private"} {
		if strings.Contains(strings.ToLower(bodyText), strings.ToLower(forbidden)) {
			t.Fatalf("private/Immich data leaked into provider body: found %q", forbidden)
		}
	}
	if !strings.Contains(bodyText, "data:image/jpeg;base64,") {
		t.Fatal("synthetic image data URL missing from provider request")
	}
	if strings.Contains(bodyText, "fixture") || strings.Contains(bodyText, ".jpg") {
		t.Fatal("fixture identifiers must not be sent to the provider")
	}

	var report struct {
		Revision       int    `json:"revision"`
		RequestedModel string `json:"requestedModel"`
		Lifecycle      string `json:"lifecycle"`
		Applicable     bool   `json:"applicable"`
		Fingerprint    string `json:"policyFingerprint"`
		Compatibility  string `json:"compatibility"`
		Observations   struct {
			Image struct {
				Status string `json:"status"`
			} `json:"image"`
			JSON struct {
				Status string `json:"status"`
			} `json:"json"`
			Strict struct {
				Status string `json:"status"`
			} `json:"strict"`
		} `json:"observations"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Revision != profile.Revision || report.RequestedModel != "gpt-5.6-sol" || report.Lifecycle != "completed" {
		t.Fatalf("unexpected report identity: %+v", report)
	}
	if report.Observations.Image.Status != "supported" || report.Observations.JSON.Status != "supported" || report.Observations.Strict.Status != "supported" {
		t.Fatalf("observations = %+v", report.Observations)
	}
	if report.Compatibility != "strict-schema sample compatible" {
		t.Fatalf("compatibility = %q", report.Compatibility)
	}
	if !report.Applicable {
		t.Fatal("supported current evidence must be applicable")
	}
	if report.Fingerprint == "" || strings.Contains(report.Fingerprint, "allowedCIDRs") || strings.Contains(report.Fingerprint, "{") {
		t.Fatalf("policy fingerprint must stay opaque: %q", report.Fingerprint)
	}
	if strings.Contains(rec.Body.String(), "proxy-secret") || strings.Contains(rec.Body.String(), "data:image/jpeg") {
		t.Fatal("credentials or image bytes leaked in report")
	}
}

func TestRejectUnauthorizedStaleOrUnapprovedTests(t *testing.T) {
	hits := &atomic.Int32{}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits.Add(1)
	}))
	proxy.Listener = listener
	proxy.Start()
	t.Cleanup(proxy.Close)
	base, err := url.Parse(proxy.URL)
	if err != nil {
		t.Fatal(err)
	}
	approvedURL := "http://codex-proxy:" + base.Port() + "/v1"
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"` + approvedURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	transport := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})

	db, handler := aiCapabilityTestHandler(t, true, policy, transport)
	secret := "secret"
	owned := mustCreateProvider(t, handler, providers.Input{
		Config:  providers.Config{Name: "Owned", BaseURL: approvedURL, Model: "gpt-5.6-sol"},
		Enabled: true,
		Secret:  &secret,
	})
	disabled := mustCreateProvider(t, handler, providers.Input{
		Config:  providers.Config{Name: "Disabled", BaseURL: approvedURL, Model: "gpt-5.6-sol"},
		Enabled: false,
		Secret:  &secret,
	})
	unapproved := mustCreateProvider(t, handler, providers.Input{
		Config:  providers.Config{Name: "Unapproved", BaseURL: "https://evil.example/v1", Model: "gpt-5.6-sol"},
		Enabled: true,
		Secret:  &secret,
	})
	if err := db.createUser(context.Background(), "other", "other@example.com", "hashed"); err != nil {
		t.Fatal(err)
	}
	foreign, err := db.createAIProvider(context.Background(), "other", "foreign", providers.Input{
		Config:  providers.Config{Name: "Foreign", BaseURL: approvedURL, Model: "gpt-5.6-sol"},
		Enabled: true,
		Secret:  &secret,
	})
	if err != nil {
		t.Fatal(err)
	}
	expiredHash := sha256.Sum256([]byte("expired-session"))
	if err := db.createSession(context.Background(), hex.EncodeToString(expiredHash[:]), testUserID, time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	disabledHandler := newAIProviderHandler(db, &Config{AIPublicOrigin: aiTestOrigin, AIProviderEgressPolicy: policy}, newAIProviderDispatcher(db, false, policy, transport))

	cases := []struct {
		name       string
		handler    http.Handler
		path       string
		body       string
		origin     string
		authCookie string
		wantStatus int
		wantCode   string
	}{
		{"missing session", handler, "/ai/providers/" + owned.ID + "/test", `{"expectedRevision":1}`, aiTestOrigin, "", 401, "UNAUTHENTICATED"},
		{"expired session", handler, "/ai/providers/" + owned.ID + "/test", `{"expectedRevision":1}`, aiTestOrigin, "expired-session", 401, "UNAUTHENTICATED"},
		{"foreign profile", handler, "/ai/providers/" + foreign.ID + "/test", `{"expectedRevision":1}`, aiTestOrigin, "ai-session", 404, "PROVIDER_NOT_FOUND"},
		{"missing profile", handler, "/ai/providers/missing/test", `{"expectedRevision":1}`, aiTestOrigin, "ai-session", 404, "PROVIDER_NOT_FOUND"},
		{"stale revision", handler, "/ai/providers/" + owned.ID + "/test", `{"expectedRevision":99}`, aiTestOrigin, "ai-session", 409, "PROVIDER_CONFLICT"},
		{"disabled profile", handler, "/ai/providers/" + disabled.ID + "/test", `{"expectedRevision":1}`, aiTestOrigin, "ai-session", 409, "PROVIDER_DISABLED"},
		{"disabled installation", disabledHandler, "/ai/providers/" + owned.ID + "/test", `{"expectedRevision":1}`, aiTestOrigin, "ai-session", 503, "AI_DISABLED"},
		{"invalid origin", handler, "/ai/providers/" + owned.ID + "/test", `{"expectedRevision":1}`, "https://attacker.example", "ai-session", 403, "ORIGIN_REJECTED"},
		{"unapproved destination", handler, "/ai/providers/" + unapproved.ID + "/test", `{"expectedRevision":1}`, aiTestOrigin, "ai-session", 403, "DESTINATION_DENIED"},
	}
	foreignBody := ""
	missingBody := ""
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := hits.Load()
			req := httptest.NewRequest("POST", tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			if tc.authCookie != "" {
				req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: tc.authCookie})
			}
			rec := httptest.NewRecorder()
			tc.handler.ServeHTTP(rec, req)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d body = %s; want %d", rec.Code, rec.Body.String(), tc.wantStatus)
			}
			var failure struct{ Code string }
			if err := json.Unmarshal(rec.Body.Bytes(), &failure); err != nil || failure.Code != tc.wantCode {
				t.Fatalf("error = %s, want code %s", rec.Body.String(), tc.wantCode)
			}
			if hits.Load() != before {
				t.Fatal("provider request made before rejection completed")
			}
			if tc.name == "foreign profile" {
				foreignBody = rec.Body.String()
			}
			if tc.name == "missing profile" {
				missingBody = rec.Body.String()
			}
		})
	}
	var foreignErr, missingErr struct {
		Code, Message string
	}
	if err := json.Unmarshal([]byte(foreignBody), &foreignErr); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(missingBody), &missingErr); err != nil {
		t.Fatal(err)
	}
	if foreignErr.Code != missingErr.Code || foreignErr.Message != missingErr.Message {
		t.Fatalf("foreign and missing must be indistinguishable: %q vs %q", foreignBody, missingBody)
	}
}

func TestRejectInjectedTestPayloadsAndOptions(t *testing.T) {
	hits := &atomic.Int32{}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits.Add(1)
	}))
	proxy.Listener = listener
	proxy.Start()
	t.Cleanup(proxy.Close)
	base, err := url.Parse(proxy.URL)
	if err != nil {
		t.Fatal(err)
	}
	approvedURL := "http://codex-proxy:" + base.Port() + "/v1"
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"` + approvedURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	db, handler := aiCapabilityTestHandler(t, true, policy, providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	}))
	secret := "secret"
	profile := mustCreateProvider(t, handler, providers.Input{
		Config:  providers.Config{Name: "Owned", BaseURL: approvedURL, Model: "gpt-5.6-sol"},
		Enabled: true,
		Secret:  &secret,
	})
	path := "/ai/providers/" + profile.ID + "/test"

	cases := []struct {
		name        string
		contentType string
		body        string
	}{
		{"unsupported fields", "application/json", `{"expectedRevision":1,"prompt":"hack","model":"other","image":"x","assetId":"a","url":"http://x","stream":true}`},
		{"malformed JSON", "application/json", `{`},
		{"non-JSON content type", "text/plain", `{"expectedRevision":1}`},
		{"oversized body", "application/json", `{"expectedRevision":1,"pad":"` + strings.Repeat("x", 2048) + `"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			beforeHits := hits.Load()
			req := httptest.NewRequest("POST", path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.contentType)
			req.Header.Set("Origin", aiTestOrigin)
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "ai-session"})
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code < 400 {
				t.Fatalf("status = %d body = %s; want rejection", rec.Code, rec.Body.String())
			}
			var failure struct{ Code string }
			if err := json.Unmarshal(rec.Body.Bytes(), &failure); err != nil || failure.Code == "" {
				t.Fatalf("unsafe error envelope: %s", rec.Body.String())
			}
			if hits.Load() != beforeHits {
				t.Fatal("external request made for rejected payload")
			}
			var stored int
			if err := db.db.QueryRow(`SELECT count(*) FROM ai_provider_capability_checks WHERE profileID = ?`, profile.ID).Scan(&stored); err != nil {
				t.Fatal(err)
			}
			if stored != 0 {
				t.Fatalf("accepted/stored %d tests after rejection", stored)
			}
		})
	}
}

func mustCreateProvider(t *testing.T, handler http.Handler, input providers.Input) aiProviderProfile {
	t.Helper()
	rec := aiRequest(handler, "POST", "/ai/providers", mustJSON(t, input), aiTestOrigin, true)
	var profile aiProviderProfile
	if err := json.Unmarshal(rec.Body.Bytes(), &profile); err != nil || rec.Code != 201 {
		t.Fatalf("create provider = %s err=%v", rec.Body.String(), err)
	}
	return profile
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
