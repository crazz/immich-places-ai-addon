package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestSendOnlyTheRevisionBoundProviderCredential(t *testing.T) {
	var received struct {
		hits                                          int
		auth, cookie, immich, contentType, accept, ua string
		headerNames                                   []string
		body                                          []byte
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received.hits++
		received.auth = r.Header.Get("Authorization")
		received.cookie = r.Header.Get("Cookie")
		received.immich = r.Header.Get("x-api-key")
		received.contentType = r.Header.Get("Content-Type")
		received.accept = r.Header.Get("Accept")
		received.ua = r.Header.Get("User-Agent")
		for name := range r.Header {
			received.headerNames = append(received.headerNames, name)
		}
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
	baseURL := "http://127.0.0.1:" + base.Port() + "/v1"
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"` + baseURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`)
	if err != nil {
		t.Fatal(err)
	}

	const providerSecret = "revision-one-provider-secret"
	db := newTestDB(t)
	ctx := context.Background()
	profile, err := db.createAIProvider(ctx, testUserID, "cred-profile", providers.Input{
		Config:  providers.Config{Name: "Cred", BaseURL: baseURL, Model: "vision"},
		Enabled: true,
		Secret:  ptr(providerSecret),
	})
	if err != nil {
		t.Fatal(err)
	}

	transport := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})
	dispatcher := newAIProviderDispatcher(db, true, policy, transport)
	_, err = dispatcher.Dispatch(ctx, providers.DispatchRequest{
		OwnerID:   testUserID,
		ProfileID: profile.ID,
		Revision:  profile.Revision,
		Body:      []byte(`{"model":"vision"}`),
	})
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
	if received.hits != 1 {
		t.Fatalf("approved destination hits=%d", received.hits)
	}
	if received.auth != "Bearer "+providerSecret {
		t.Fatalf("authorization = %q", received.auth)
	}
	if received.cookie != "" || received.immich != "" {
		t.Fatalf("browser or Immich credential leaked: cookie=%q immich=%q", received.cookie, received.immich)
	}
	if received.contentType != "application/json" || received.accept != "application/json" || received.ua == "" {
		t.Fatalf("permitted transport headers missing: %+v", received)
	}
	for _, name := range received.headerNames {
		switch strings.ToLower(name) {
		case "authorization", "content-type", "accept", "user-agent", "content-length", "host", "connection":
			continue
		default:
			t.Fatalf("unexpected outbound header %q", name)
		}
	}
}

func TestCorruptOrRemovedCredentialsCannotBeReused(t *testing.T) {
	hits := 0
	var lastAuth string
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		lastAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)

	base, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	baseURL := "http://127.0.0.1:" + base.Port() + "/v1"
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"` + baseURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	transport := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})
	db := newTestDB(t)
	ctx := context.Background()
	dispatcher := newAIProviderDispatcher(db, true, policy, transport)

	corrupt, err := db.createAIProvider(ctx, testUserID, "corrupt", providers.Input{
		Config:  providers.Config{Name: "Corrupt", BaseURL: baseURL, Model: "vision"},
		Enabled: true,
		Secret:  ptr("good-secret"),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, stored := range []string{"plaintext-secret", "enc:broken"} {
		hits = 0
		if _, err := db.db.ExecContext(ctx, `UPDATE ai_provider_versions SET secretCiphertext=? WHERE userID=? AND profileID=?`, stored, testUserID, corrupt.ID); err != nil {
			t.Fatal(err)
		}
		_, err = dispatcher.Dispatch(ctx, providers.DispatchRequest{
			OwnerID: testUserID, ProfileID: corrupt.ID, Revision: corrupt.Revision, Body: []byte(`{}`),
		})
		if err == nil || !errors.Is(err, providers.ErrCredential) || hits != 0 {
			t.Fatalf("stored=%q hits=%d err=%v", stored, hits, err)
		}
	}

	withSecret, err := db.createAIProvider(ctx, testUserID, "cleared", providers.Input{
		Config:  providers.Config{Name: "Cleared", BaseURL: baseURL, Model: "vision"},
		Enabled: true,
		Secret:  ptr("must-not-leak"),
	})
	if err != nil {
		t.Fatal(err)
	}
	clearInput := providers.Input{
		Config:  providers.Config{Name: "Cleared", BaseURL: baseURL, Model: "vision"},
		Enabled: true,
		Secret:  ptr(""),
	}
	cleared, err := db.updateAIProvider(ctx, testUserID, withSecret.ID, withSecret.Revision, clearInput)
	if err != nil {
		t.Fatal(err)
	}
	hits, lastAuth = 0, "stale"
	result, err := dispatcher.Dispatch(ctx, providers.DispatchRequest{
		OwnerID: testUserID, ProfileID: withSecret.ID, Revision: withSecret.Revision, Body: []byte(`{}`),
	})
	if err != nil || hits != 1 || lastAuth != "" || result.StatusCode != http.StatusOK {
		t.Fatalf("removed secret still transmitted: hits=%d auth=%q result=%+v err=%v", hits, lastAuth, result, err)
	}
	hits, lastAuth = 0, "stale"
	result, err = dispatcher.Dispatch(ctx, providers.DispatchRequest{
		OwnerID: testUserID, ProfileID: cleared.ID, Revision: cleared.Revision, Body: []byte(`{}`),
	})
	if err != nil || hits != 1 || lastAuth != "" {
		t.Fatalf("credential-free current revision failed: hits=%d auth=%q err=%v", hits, lastAuth, err)
	}

	free, err := db.createAIProvider(ctx, testUserID, "free", providers.Input{
		Config:  providers.Config{Name: "Free", BaseURL: baseURL, Model: "vision"},
		Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	hits, lastAuth = 0, "stale"
	_, err = dispatcher.Dispatch(ctx, providers.DispatchRequest{
		OwnerID: testUserID, ProfileID: free.ID, Revision: free.Revision, Body: []byte(`{}`),
	})
	if err != nil || hits != 1 || lastAuth != "" {
		t.Fatalf("deliberately credential-free dispatch: hits=%d auth=%q err=%v", hits, lastAuth, err)
	}
}

func TestForbiddenCredentialHeaderCharactersFailClosed(t *testing.T) {
	hits := 0
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits++ }))
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)
	base, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	baseURL := "http://127.0.0.1:" + base.Port() + "/v1"
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"` + baseURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	db := newTestDB(t)
	ctx := context.Background()
	profile, err := db.createAIProvider(ctx, testUserID, "hdr", providers.Input{
		Config:  providers.Config{Name: "Hdr", BaseURL: baseURL, Model: "vision"},
		Enabled: true,
		Secret:  ptr("ok"),
	})
	if err != nil {
		t.Fatal(err)
	}
	cipher, err := encryptValue(db.encryptionKey, "bad\r\nsecret")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.db.ExecContext(ctx, `UPDATE ai_provider_versions SET secretCiphertext=? WHERE userID=? AND profileID=?`, cipher, testUserID, profile.ID); err != nil {
		t.Fatal(err)
	}
	dispatcher := newAIProviderDispatcher(db, true, policy, providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	}))
	_, err = dispatcher.Dispatch(ctx, providers.DispatchRequest{
		OwnerID: testUserID, ProfileID: profile.ID, Revision: profile.Revision, Body: []byte(`{}`),
	})
	if err == nil || !errors.Is(err, providers.ErrCredential) || hits != 0 {
		t.Fatalf("hits=%d err=%v", hits, err)
	}
}

func TestPlaceholderLocalCredentialIsProviderAuthorization(t *testing.T) {
	var auth string
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)
	base, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	baseURL := "http://127.0.0.1:" + base.Port() + "/v1"
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"` + baseURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	db := newTestDB(t)
	ctx := context.Background()
	profile, err := db.createAIProvider(ctx, testUserID, "local-secret", providers.Input{
		Config:  providers.Config{Name: "Local", BaseURL: baseURL, Model: "vision"},
		Enabled: true,
		Secret:  ptr("local"),
	})
	if err != nil {
		t.Fatal(err)
	}
	dispatcher := newAIProviderDispatcher(db, true, policy, providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	}))
	_, err = dispatcher.Dispatch(ctx, providers.DispatchRequest{
		OwnerID: testUserID, ProfileID: profile.ID, Revision: profile.Revision, Body: []byte(`{}`),
	})
	if err != nil || auth != "Bearer local" {
		t.Fatalf("auth=%q err=%v", auth, err)
	}
}
