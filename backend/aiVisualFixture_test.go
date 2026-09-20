package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

type aiVisualFixture struct {
	test               *testing.T
	image              *aiImageFixture
	analyzer           *aiVisualAnalyzer
	request            analysis.Request
	provider           *httptest.Server
	hits, reservations atomic.Int32
	handle             func(http.ResponseWriter, *http.Request) bool
}

func newAIVisualFixture(t *testing.T) *aiVisualFixture {
	t.Helper()
	f := &aiVisualFixture{image: newAIImageFixture(t), test: t}
	content, err := os.ReadFile("internal/ai/results/testdata/ai-analysis-result.unknown.json")
	if err != nil {
		t.Fatal(err)
	}
	f.provider = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.hits.Add(1)
		body, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			t.Error(readErr)
			return
		}
		var request struct {
			Model string `json:"model"`
		}
		if json.Unmarshal(body, &request) != nil || request.Model != "bound-model" {
			t.Error("wrong bound wire model")
		}
		if r.Method != "POST" || r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer provider-secret" || r.Header.Get("x-api-key") != "" {
			t.Error("wrong provider authority or path")
		}
		for _, private := range []string{testUserID, selectionA, f.image.store.binding, "private-immich-image-key", f.image.server.URL} {
			if strings.Contains(string(body), private) {
				t.Error("private source in provider body")
			}
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		if f.handle != nil && f.handle(w, r) {
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": string(content)}}}})
	}))
	t.Cleanup(f.provider.Close)
	policyJSON, _ := json.Marshal([]map[string]any{{"baseURL": f.provider.URL + "/v1", "addressClass": "local", "allowedCIDRs": []string{"127.0.0.0/8"}}})
	policy, err := providers.ParseEgressPolicy(string(policyJSON))
	if err != nil {
		t.Fatal(err)
	}
	secret := "provider-secret"
	_, err = f.image.db.createAIProvider(context.Background(), testUserID, "visual-profile", providers.Input{Config: providers.Config{Name: "Visual", BaseURL: f.provider.URL + "/v1", Model: "bound-model"}, Enabled: true, Secret: &secret})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	admission, err := f.image.db.admitAIProviderCapability(context.Background(), testUserID, "visual-profile", 1, policyFingerprint(policy), now, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	obs := capabilities.Observations{Image: capabilities.Observation{Status: capabilities.StatusSupported}, JSON: capabilities.Observation{Status: capabilities.StatusSupported}, Strict: capabilities.Observation{Status: capabilities.StatusSupported}}
	report := capabilities.Report{AttemptID: admission.AttemptID, ProfileID: admission.ProfileID, Revision: 1, ProtocolVersion: capabilities.ProtocolVersion, PolicyFingerprint: admission.PolicyFingerprint, Lifecycle: "completed", StartedAt: now, DeadlineAt: now.Add(time.Minute), CompletedAt: &now, RequestedModel: "bound-model", Observations: obs, Compatibility: capabilities.CompatibilitySummary(obs), Applicable: true}
	if err := f.image.db.completeAIProviderCapability(context.Background(), testUserID, report); err != nil {
		t.Fatal(err)
	}
	prepared, err := f.image.service.prepare(context.Background(), testUserID, f.image.store.binding, selectionA)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(prepared.Release)
	dispatcher := newAIProviderDispatcher(f.image.db, true, policy, providerhttp.New(providerhttp.Options{}))
	f.analyzer, err = newAIVisualAnalyzer(f.image.db, f.image.service, dispatcher)
	if err != nil {
		t.Fatal(err)
	}
	f.request = analysis.Request{Owner: testUserID, Installation: f.image.store.binding, Asset: selectionA, ProfileID: "visual-profile", Revision: 1, Format: "strict", Languages: []string{"en"}, PrimaryLanguage: "en", Image: prepared, Guard: analysis.Guard{Authorize: func(context.Context) error { return nil }, Reserve: func(context.Context) error { f.reservations.Add(1); return nil }}}
	return f
}

func (f *aiVisualFixture) exec(query string) {
	f.test.Helper()
	if _, err := f.image.db.db.Exec(query); err != nil {
		f.test.Fatal(err)
	}
}
