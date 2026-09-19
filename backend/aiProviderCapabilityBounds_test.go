package main

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestCapabilityProbeResponseIsBoundedBelowGeneralTransportDefault(t *testing.T) {
	padded := []byte(`{"id":"cmpl-1","model":"gpt-5.6-sol","choices":[{"message":{"role":"assistant","content":"color: blue shape: circle ` +
		strings.Repeat("x", providerhttp.MaxCapabilityResponseBytes) + `"},"finish_reason":"stop"}]}`)
	if len(padded) <= providerhttp.MaxCapabilityResponseBytes || len(padded) >= 1<<20 {
		t.Fatalf("fixture response of %d bytes must exceed only the capability ceiling", len(padded))
	}
	var hits atomic.Int32
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(padded)
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
	transport := newAIProviderCapabilityTransport(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})
	_, handler := aiCapabilityTestHandler(t, true, policy, transport)
	secret := "secret"
	profile := mustCreateProvider(t, handler, providers.Input{
		Config:  providers.Config{Name: "Owned", BaseURL: approvedURL, Model: "gpt-5.6-sol"},
		Enabled: true,
		Secret:  &secret,
	})

	rec := aiRequest(handler, "POST", "/ai/providers/"+profile.ID+"/test", mustJSON(t, map[string]any{"expectedRevision": 1}), aiTestOrigin, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s, want the persisted report on 200", rec.Code, rec.Body.String())
	}
	var report capabilities.Report
	if err := json.Unmarshal(rec.Body.Bytes(), &report); err != nil {
		t.Fatalf("report = %s err=%v", rec.Body.String(), err)
	}
	if report.Observations.Image.Status != capabilities.StatusUnverified || report.Observations.Image.Reason != "provider_failure" {
		t.Fatalf("image observation = %+v, want the oversize response rejected", report.Observations.Image)
	}
	if hits.Load() != 1 {
		t.Fatalf("provider requests = %d, want the sequence stopped after the bounded rejection", hits.Load())
	}
}
