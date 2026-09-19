package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"immich-places-backend/internal/ai/providers"
)

func TestSavedSettingsDoNotApproveDestination(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits++
	}))
	t.Cleanup(server.Close)

	db := newTestDB(t)
	ctx := context.Background()
	input := providers.Input{
		Config:  providers.Config{Name: "Unapproved", BaseURL: server.URL + "/v1", Model: "vision"},
		Enabled: true,
	}
	profile, err := db.createAIProvider(ctx, testUserID, "profile", input)
	if err != nil || hits != 0 {
		t.Fatalf("save must succeed offline: profile=%+v hits=%d err=%v", profile, hits, err)
	}

	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"https://other.example/v1","addressClass":"public"}]`)
	if err != nil {
		t.Fatal(err)
	}
	dispatcher := newAIProviderDispatcher(db, true, policy, nil)
	_, err = dispatcher.Dispatch(ctx, providers.DispatchRequest{
		OwnerID:   testUserID,
		ProfileID: profile.ID,
		Revision:  profile.Revision,
		Body:      []byte(`{}`),
	})
	if err == nil || !providers.IsPolicyError(err) {
		t.Fatalf("dispatch must fail with policy error before contact: %v", err)
	}
	if hits != 0 {
		t.Fatalf("unapproved destination received %d requests", hits)
	}

	empty := newAIProviderDispatcher(db, true, providers.EgressPolicy{}, nil)
	_, err = empty.Dispatch(ctx, providers.DispatchRequest{
		OwnerID:   testUserID,
		ProfileID: profile.ID,
		Revision:  profile.Revision,
		Body:      []byte(`{}`),
	})
	if err == nil || !providers.IsPolicyError(err) || hits != 0 {
		t.Fatalf("empty policy must deny without contact: hits=%d err=%v", hits, err)
	}
}
