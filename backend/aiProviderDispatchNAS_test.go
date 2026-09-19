package main

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestUseExistingApprovedNASProxy(t *testing.T) {
	var received struct {
		method, path, auth, contentType, accept, ua string
		body                                        []byte
		hits                                        int
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received.hits++
		received.method = r.Method
		received.path = r.URL.Path
		received.auth = r.Header.Get("Authorization")
		received.contentType = r.Header.Get("Content-Type")
		received.accept = r.Header.Get("Accept")
		received.ua = r.Header.Get("User-Agent")
		received.body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)

	base, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	baseURL := "http://codex-proxy:" + base.Port() + "/v1"
	policyJSON := `[{"baseURL":"` + baseURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`
	policy, err := providers.ParseEgressPolicy(policyJSON)
	if err != nil {
		t.Fatal(err)
	}

	db := newTestDB(t)
	ctx := context.Background()
	profile, err := db.createAIProvider(ctx, testUserID, "nas", providers.Input{
		Config:  providers.Config{Name: "NAS", BaseURL: baseURL, Model: "gpt-5.6-luna"},
		Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	transport := providerhttp.New(providerhttp.Options{
		LookupIP: func(ctx context.Context, host string) ([]net.IP, error) {
			if host != "codex-proxy" {
				t.Fatalf("unexpected host lookup: %s", host)
			}
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})
	dispatcher := newAIProviderDispatcher(db, true, policy, transport)
	result, err := dispatcher.Dispatch(ctx, providers.DispatchRequest{
		OwnerID:   testUserID,
		ProfileID: profile.ID,
		Revision:  profile.Revision,
		Body:      []byte(`{"model":"gpt-5.6-luna"}`),
	})
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if received.hits != 1 || received.method != http.MethodPost || received.path != "/v1/chat/completions" {
		t.Fatalf("approved endpoint not contacted exactly once: %+v", received)
	}
	if received.auth != "" || received.contentType != "application/json" || received.accept != "application/json" || received.ua == "" {
		t.Fatalf("unexpected headers: %+v", received)
	}
	var payload map[string]any
	if err := json.Unmarshal(result.Body, &payload); err != nil || payload["ok"] != true {
		t.Fatalf("response = %s err=%v", result.Body, err)
	}
}
