package main

import (
	"context"
	"errors"
	"net"
	"testing"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestDenyMissingForeignOrDisabledProfiles(t *testing.T) {
	lookups := 0
	transport := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			lookups++
			t.Fatal("DNS must not run for unavailable profiles")
			return nil, nil
		},
	})
	db := newTestDB(t)
	ctx := context.Background()
	if err := db.createUser(ctx, "other-user", "other@example.com", "hashed"); err != nil {
		t.Fatal(err)
	}
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"http://127.0.0.1:9/v1","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	owned, err := db.createAIProvider(ctx, testUserID, "owned", providers.Input{
		Config:  providers.Config{Name: "Owned", BaseURL: "http://127.0.0.1:9/v1", Model: "vision"},
		Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := db.createAIProvider(ctx, "other-user", "foreign", providers.Input{
		Config:  providers.Config{Name: "Foreign", BaseURL: "http://127.0.0.1:9/v1", Model: "vision"},
		Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	disabled, err := db.createAIProvider(ctx, testUserID, "disabled", providers.Input{
		Config:  providers.Config{Name: "Disabled", BaseURL: "http://127.0.0.1:9/v1", Model: "vision"},
		Enabled: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	enabledDispatcher := newAIProviderDispatcher(db, true, policy, transport)
	disabledAI := newAIProviderDispatcher(db, false, policy, transport)

	missing := mustDispatchError(t, enabledDispatcher, providers.DispatchRequest{OwnerID: testUserID, ProfileID: "missing", Revision: 1, Body: []byte(`{}`)})
	if !errors.Is(missing, providers.ErrUnavailable) {
		t.Fatalf("missing profile: %v", missing)
	}
	foreignErr := mustDispatchError(t, enabledDispatcher, providers.DispatchRequest{OwnerID: testUserID, ProfileID: foreign.ID, Revision: foreign.Revision, Body: []byte(`{}`)})
	if !errors.Is(foreignErr, providers.ErrUnavailable) || foreignErr.Error() != missing.Error() {
		t.Fatalf("foreign and missing must be indistinguishable: missing=%v foreign=%v", missing, foreignErr)
	}
	if !errors.Is(mustDispatchError(t, enabledDispatcher, providers.DispatchRequest{OwnerID: testUserID, ProfileID: owned.ID, Revision: 99, Body: []byte(`{}`)}), providers.ErrUnavailable) {
		t.Fatal("missing revision must be unavailable")
	}
	if !errors.Is(mustDispatchError(t, enabledDispatcher, providers.DispatchRequest{OwnerID: testUserID, ProfileID: disabled.ID, Revision: disabled.Revision, Body: []byte(`{}`)}), providers.ErrDisabled) {
		t.Fatal("disabled profile must be rejected")
	}
	if !errors.Is(mustDispatchError(t, disabledAI, providers.DispatchRequest{OwnerID: testUserID, ProfileID: owned.ID, Revision: owned.Revision, Body: []byte(`{}`)}), providers.ErrDisabled) {
		t.Fatal("disabled AI must be rejected")
	}
	if _, err := db.db.ExecContext(ctx, `DELETE FROM users WHERE ID=?`, testUserID); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(mustDispatchError(t, enabledDispatcher, providers.DispatchRequest{OwnerID: testUserID, ProfileID: owned.ID, Revision: owned.Revision, Body: []byte(`{}`)}), providers.ErrUnavailable) {
		t.Fatal("deleted owner must be unavailable")
	}
	if lookups != 0 {
		t.Fatalf("DNS lookups occurred: %d", lookups)
	}
}

func TestDisablementAppliesAcrossHistoricalRevisions(t *testing.T) {
	lookups := 0
	transport := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			lookups++
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})
	db := newTestDB(t)
	ctx := context.Background()
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"http://127.0.0.1:9/v1","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	first, err := db.createAIProvider(ctx, testUserID, "hist", providers.Input{
		Config:  providers.Config{Name: "Hist", BaseURL: "http://127.0.0.1:9/v1", Model: "v1"},
		Enabled: true,
		Secret:  ptr("s1"),
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := db.updateAIProvider(ctx, testUserID, first.ID, first.Revision, providers.Input{
		Config:  providers.Config{Name: "Hist", BaseURL: "http://127.0.0.1:9/v1", Model: "v2"},
		Enabled: true,
		Secret:  ptr("s2"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.db.ExecContext(ctx, `UPDATE ai_provider_profiles SET enabled=0 WHERE userID=? AND id=?`, testUserID, first.ID); err != nil {
		t.Fatal(err)
	}
	dispatcher := newAIProviderDispatcher(db, true, policy, transport)
	for _, revision := range []int{first.Revision, second.Revision} {
		if !errors.Is(mustDispatchError(t, dispatcher, providers.DispatchRequest{
			OwnerID: testUserID, ProfileID: first.ID, Revision: revision, Body: []byte(`{}`),
		}), providers.ErrDisabled) {
			t.Fatalf("revision %d must be disabled", revision)
		}
	}
	if lookups != 0 {
		t.Fatalf("DNS lookups occurred: %d", lookups)
	}
}

func mustDispatchError(t *testing.T, d *providers.Dispatcher, req providers.DispatchRequest) error {
	t.Helper()
	_, err := d.Dispatch(context.Background(), req)
	if err == nil {
		t.Fatal("expected dispatch error")
	}
	return err
}
