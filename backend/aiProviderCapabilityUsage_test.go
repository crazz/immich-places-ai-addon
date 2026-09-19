package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestCapabilityUsageIsAggregatedAndPersisted(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		content := "color: blue\nshape: circle"
		if hits.Add(1) > 1 {
			content = `{"color":"blue","shape":"circle"}`
		}
		encoded, _ := json.Marshal(content)
		fmt.Fprintf(w, `{"usage":{"prompt_tokens":10,"completion_tokens":2,"total_tokens":12},"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":%s}}]}`, encoded)
	}))
	t.Cleanup(server.Close)
	endpoint, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	baseURL := "http://provider.test:" + endpoint.Port() + "/v1"
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"` + baseURL + `","addressClass":"local","allowedCIDRs":["127.0.0.1/32"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	transport := providerhttp.New(providerhttp.Options{LookupIP: func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("127.0.0.1")}, nil
	}})
	db, handler := aiCapabilityTestHandler(t, true, policy, transport)
	profile := mustCreateProvider(t, handler, providers.Input{
		Config: providers.Config{Name: "Owned", BaseURL: baseURL, Model: "vision"}, Enabled: true,
	})
	response := aiRequest(handler, "POST", "/ai/providers/"+profile.ID+"/test", `{"expectedRevision":1}`, aiTestOrigin, true)
	var report capabilities.Report
	if err := json.Unmarshal(response.Body.Bytes(), &report); err != nil || response.Code != http.StatusOK || hits.Load() != 3 {
		t.Fatalf("test response = %d %s; hits=%d err=%v", response.Code, response.Body.String(), hits.Load(), err)
	}
	assertCapabilityUsage(t, report.Usage)
	stored, err := db.loadAIProviderCapability(context.Background(), testUserID, profile.ID, 1, capabilities.ApplicabilityContext{})
	if err != nil || stored == nil {
		t.Fatalf("reload report: %+v %v", stored, err)
	}
	assertCapabilityUsage(t, stored.Usage)
	if !stored.InputMayBeConsumed || hits.Load() != 3 {
		t.Fatal("reload lost the usage disclosure or repeated provider calls")
	}
}

func assertCapabilityUsage(t *testing.T, usage *capabilities.Usage) {
	t.Helper()
	if usage == nil || usage.PromptTokens == nil || *usage.PromptTokens != 30 ||
		usage.CompletionTokens == nil || *usage.CompletionTokens != 6 ||
		usage.TotalTokens == nil || *usage.TotalTokens != 36 {
		t.Fatalf("expected exact three-probe usage, got %+v", usage)
	}
}
