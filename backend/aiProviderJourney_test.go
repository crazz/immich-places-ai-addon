package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestAIProviderHTTPJourneyMakesNoExternalCalls(t *testing.T) {
	var calls atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.WriteHeader(500) }))
	defer provider.Close()
	db, handler := aiTestHandler(t, true)
	input := providerInput()
	input.BaseURL = provider.URL
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	created := aiRequest(handler, "POST", "/ai/providers", string(body), aiTestOrigin, true)
	var profile aiProviderProfile
	if err := json.Unmarshal(created.Body.Bytes(), &profile); err != nil || created.Code != 201 {
		t.Fatalf("create = %s, error = %v", created.Body.String(), err)
	}
	input.Enabled, input.Secret, input.Name = false, nil, "Disabled"
	edit, err := json.Marshal(aiProviderRequest{Input: input, ExpectedRevision: 1})
	if err != nil {
		t.Fatal(err)
	}
	saved := aiRequest(handler, "PUT", "/ai/providers/"+profile.ID, string(edit), aiTestOrigin, true)
	if saved.Code != 200 {
		t.Fatalf("save = %d: %s", saved.Code, saved.Body.String())
	}
	stale := aiRequest(handler, "PUT", "/ai/providers/"+profile.ID, string(edit), aiTestOrigin, true)
	if stale.Code != 409 || !strings.Contains(stale.Body.String(), "PROVIDER_CONFLICT") {
		t.Fatalf("stale = %d: %s", stale.Code, stale.Body.String())
	}
	if err := db.createUser(context.Background(), "other", "other@example.com", "hashed"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.createAIProvider(context.Background(), "other", "private-other", providerInput()); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"private-other", "missing"} {
		response := aiRequest(handler, "PUT", "/ai/providers/"+id, string(edit), aiTestOrigin, true)
		if response.Code != 404 || !strings.Contains(response.Body.String(), "PROVIDER_NOT_FOUND") {
			t.Fatalf("foreign/missing ID = %d: %s", response.Code, response.Body.String())
		}
	}
	listed := aiRequest(handler, "GET", "/ai/providers", "", "", true)
	var list struct {
		Enabled bool
		Items   []aiProviderProfile
	}
	if err := json.Unmarshal(listed.Body.Bytes(), &list); err != nil || listed.Code != 200 || !list.Enabled || len(list.Items) != 1 || list.Items[0].Enabled || list.Items[0].Revision != 2 {
		t.Fatalf("list = %s, error = %v", listed.Body.String(), err)
	}
	disabledHandler := newAIProviderHandler(db, &Config{}, nil)
	disabled := aiRequest(disabledHandler, "GET", "/ai/providers", "", "", true)
	if err := json.Unmarshal(disabled.Body.Bytes(), &list); err != nil || list.Enabled || len(list.Items) != 0 {
		t.Fatalf("disabled list exposed stored profiles: %s", disabled.Body.String())
	}
	if calls.Load() != 0 {
		t.Fatalf("profile operations made %d external requests", calls.Load())
	}
}
