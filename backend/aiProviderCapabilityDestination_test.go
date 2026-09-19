package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestUnapprovedDestinationIsDeniedBeforeAcceptingATest(t *testing.T) {
	var hits atomic.Int32
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capabilityFixtureResponse(w, r, &hits)
	}))
	proxy.Listener = listener
	proxy.Start()
	t.Cleanup(proxy.Close)
	base, err := url.Parse(proxy.URL)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"http://codex-proxy:` + base.Port() + `/v1","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`)
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
		Config:  providers.Config{Name: "Unapproved", BaseURL: "https://evil.example/v1", Model: "gpt-5.6-sol"},
		Enabled: true,
		Secret:  &secret,
	})

	rec := aiRequest(handler, "POST", "/ai/providers/"+profile.ID+"/test", mustJSON(t, map[string]any{"expectedRevision": 1}), aiTestOrigin, true)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d body = %s, want 403", rec.Code, rec.Body.String())
	}
	var failure struct{ Code string }
	if err := json.Unmarshal(rec.Body.Bytes(), &failure); err != nil || failure.Code != "DESTINATION_DENIED" {
		t.Fatalf("error = %s, want DESTINATION_DENIED", rec.Body.String())
	}
	if hits.Load() != 0 {
		t.Fatalf("provider requests = %d, want none", hits.Load())
	}
	var stored int
	if err := db.db.QueryRow(`SELECT count(*) FROM ai_provider_capability_checks WHERE profileID = ?`, profile.ID).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored != 0 {
		t.Fatalf("accepted/stored %d tests for an unapproved destination", stored)
	}
}
