package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestSettingsAndExistingWorkflowsStayOfflineFromProviders(t *testing.T) {
	var providerHits atomic.Int64
	provider := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		providerHits.Add(1)
	}))
	t.Cleanup(provider.Close)

	db, handler := aiTestHandler(t, true)
	input := providerInput()
	input.BaseURL = provider.URL
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}

	created := aiRequest(handler, "POST", "/ai/providers", string(body), aiTestOrigin, true)
	if created.Code != http.StatusCreated {
		t.Fatalf("create = %d: %s", created.Code, created.Body.String())
	}
	var profile aiProviderProfile
	if err := json.Unmarshal(created.Body.Bytes(), &profile); err != nil {
		t.Fatal(err)
	}

	listed := aiRequest(handler, "GET", "/ai/providers", "", "", true)
	if listed.Code != http.StatusOK {
		t.Fatalf("list = %d: %s", listed.Code, listed.Body.String())
	}

	input.Name = "Edited offline"
	edit, err := json.Marshal(aiProviderRequest{Input: input, ExpectedRevision: profile.Revision})
	if err != nil {
		t.Fatal(err)
	}
	saved := aiRequest(handler, "PUT", "/ai/providers/"+profile.ID, string(edit), aiTestOrigin, true)
	if saved.Code != http.StatusOK {
		t.Fatalf("edit = %d: %s", saved.Code, saved.Body.String())
	}

	input.Enabled = false
	disableBody, err := json.Marshal(aiProviderRequest{Input: input, ExpectedRevision: profile.Revision + 1})
	if err != nil {
		t.Fatal(err)
	}
	disabled := aiRequest(handler, "PUT", "/ai/providers/"+profile.ID, string(disableBody), aiTestOrigin, true)
	if disabled.Code != http.StatusOK {
		t.Fatalf("disable = %d: %s", disabled.Code, disabled.Body.String())
	}

	for _, route := range []struct {
		method, path string
	}{
		{http.MethodPost, "/ai/dispatch"},
		{http.MethodPost, "/ai/providers/dispatch"},
		{http.MethodPost, "/ai/provider-dispatch"},
		{http.MethodGet, "/ai/dispatch"},
	} {
		rec := aiRequest(handler, route.method, route.path, `{}`, aiTestOrigin, true)
		if rec.Code != http.StatusNotFound && rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("public dispatch route available: %s %s -> %d %s", route.method, route.path, rec.Code, rec.Body.String())
		}
	}

	disabledAI := newAIProviderHandler(db, &Config{}, nil)
	disabledList := aiRequest(disabledAI, "GET", "/ai/providers", "", "", true)
	var list struct {
		Enabled bool
		Items   []aiProviderProfile
	}
	if err := json.Unmarshal(disabledList.Body.Bytes(), &list); err != nil || list.Enabled || len(list.Items) != 0 {
		t.Fatalf("AI-disabled listing changed: %s", disabledList.Body.String())
	}

	if providerHits.Load() != 0 {
		t.Fatalf("profile operations contacted provider %d times", providerHits.Load())
	}
	if strings.Contains(created.Body.String(), "sk-") {
		t.Fatalf("create response leaked secret material")
	}
}
