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

	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestCapabilityTestStopsOnAuthFailureWithoutFallback(t *testing.T) {
	assertCapabilityStopsOnProbeFailure(t, probeFailureCase{
		name:       "authentication",
		failStatus: http.StatusUnauthorized,
		failBody:   `{"error":{"message":"nope","code":"invalid_api_key"}}`,
		wantReason: "authentication",
		wantHTTP:   http.StatusOK,
	})
}

func TestCapabilityTestStopsOnUnavailableModelWithoutFallback(t *testing.T) {
	assertCapabilityStopsOnProbeFailure(t, probeFailureCase{
		name:       "unavailable_model",
		failStatus: http.StatusNotFound,
		failBody:   `{"error":{"message":"missing","code":"model_not_found"}}`,
		wantReason: "unavailable_model",
		wantHTTP:   http.StatusOK,
	})
}

func TestCapabilityTestStopsOnRateLimitWithoutFallback(t *testing.T) {
	assertCapabilityStopsOnProbeFailure(t, probeFailureCase{
		name:       "rate_limit",
		failStatus: http.StatusTooManyRequests,
		failBody:   `{"error":{"message":"slow","code":"rate_limit_exceeded"}}`,
		wantReason: "rate_limit",
		wantHTTP:   http.StatusOK,
	})
}

func TestCapabilityTestStopsOnServerFailureWithoutFallback(t *testing.T) {
	assertCapabilityStopsOnProbeFailure(t, probeFailureCase{
		name:       "server",
		failStatus: http.StatusBadGateway,
		failBody:   `{"error":{"message":"upstream"}}`,
		wantReason: "server",
		wantHTTP:   http.StatusOK,
	})
}

func TestCapabilityTestStopsOnNetworkFailureWithoutFallback(t *testing.T) {
	assertCapabilityStopsOnProbeFailure(t, probeFailureCase{
		name:            "network",
		wantReason:      "network",
		wantHTTP:        http.StatusOK,
		failOnTransmit:  true,
		transmitFailure: providers.NewTransportFailure(providers.FailureConnection, "req", 0, nil),
	})
}

func TestCapabilityTestStopsOnPolicyFailureWithoutFallback(t *testing.T) {
	assertCapabilityStopsOnProbeFailure(t, probeFailureCase{
		name:            "policy",
		wantReason:      "policy",
		wantHTTP:        http.StatusOK,
		failOnTransmit:  true,
		transmitFailure: providers.ErrPolicyDenied,
		reloadReport:    true,
	})
}

func TestCapabilityTestGeneric400IsNotServerOrUnsupported(t *testing.T) {
	assertCapabilityStopsOnProbeFailure(t, probeFailureCase{
		name:          "unknown_upstream",
		failStatus:    http.StatusBadRequest,
		failBody:      `{"error":{"message":"free-form rejection without structured code"}}`,
		wantReason:    "provider_failure",
		wantHTTP:      http.StatusOK,
		reloadReport:  true,
		forbidReasons: []string{"server", "unsupported_mode", "authentication"},
	})
}

func TestCapabilityTestReturnsPersistedPolicyFailureReport(t *testing.T) {
	var hits atomic.Int32
	var transmitCalls atomic.Int32
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
	approvedURL := "http://codex-proxy:" + base.Port() + "/v1"
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"` + approvedURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	inner := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})
	transport := &failAfterTransport{inner: inner, failAfter: 1, failErr: providers.ErrPolicyDenied, calls: &transmitCalls}
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
	if report.Lifecycle != "failed" || report.Compatibility != "failed" {
		t.Fatalf("lifecycle = %q compatibility = %q, want failed", report.Lifecycle, report.Compatibility)
	}
	if report.Observations.Image.Status != capabilities.StatusSupported {
		t.Fatalf("image observation = %+v, want the completed evidence", report.Observations.Image)
	}
	if report.Observations.JSON.Reason != "policy" || report.Observations.JSON.Status != capabilities.StatusUnverified {
		t.Fatalf("json observation = %+v, want unverified policy", report.Observations.JSON)
	}
	if report.AttemptID == "" || report.CompletedAt == nil {
		t.Fatalf("report must identify the persisted attempt: %+v", report)
	}
}

type probeFailureCase struct {
	name            string
	failStatus      int
	failBody        string
	wantReason      string
	wantHTTP        int
	failOnTransmit  bool
	transmitFailure error
	reloadReport    bool
	forbidReasons   []string
}

type failAfterTransport struct {
	inner     providers.Transport
	failAfter int32
	failErr   error
	calls     *atomic.Int32
}

func (t *failAfterTransport) Pin(ctx context.Context, rule providers.EgressRule) (providers.PinnedDestination, error) {
	return t.inner.Pin(ctx, rule)
}

func (t *failAfterTransport) Transmit(ctx context.Context, dest providers.PinnedDestination, body []byte, authorization string) (providers.DispatchResult, error) {
	n := t.calls.Add(1)
	if n > t.failAfter {
		return providers.DispatchResult{}, t.failErr
	}
	return t.inner.Transmit(ctx, dest, body, authorization)
}

func assertCapabilityStopsOnProbeFailure(t *testing.T, tc probeFailureCase) {
	t.Helper()
	var hits atomic.Int32
	var transmitCalls atomic.Int32
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n == 1 {
			capabilityFixtureResponse(w, r, nil)
			return
		}
		if tc.failOnTransmit {
			t.Fatalf("%s: HTTP fixture must not receive probe after transmit failure", tc.name)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(tc.failStatus)
		_, _ = w.Write([]byte(tc.failBody))
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
	inner := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})
	var transport providers.Transport = inner
	if tc.failOnTransmit {
		transport = &failAfterTransport{inner: inner, failAfter: 1, failErr: tc.transmitFailure, calls: &transmitCalls}
	}
	_, handler := aiCapabilityTestHandler(t, true, policy, transport)
	secret := "secret"
	profile := mustCreateProvider(t, handler, providers.Input{
		Config:  providers.Config{Name: "Owned", BaseURL: approvedURL, Model: "gpt-5.6-sol"},
		Enabled: true,
		Secret:  &secret,
	})
	rec := aiRequest(handler, "POST", "/ai/providers/"+profile.ID+"/test", mustJSON(t, map[string]any{"expectedRevision": 1}), aiTestOrigin, true)
	if rec.Code != tc.wantHTTP {
		t.Fatalf("%s status = %d body = %s, want %d", tc.name, rec.Code, rec.Body.String(), tc.wantHTTP)
	}
	wantHits := int32(2)
	if tc.failOnTransmit {
		wantHits = 1
		if transmitCalls.Load() != 2 {
			t.Fatalf("%s transmit calls = %d, want 2", tc.name, transmitCalls.Load())
		}
	}
	if hits.Load() != wantHits {
		t.Fatalf("%s fixture hits = %d, want %d (remaining probes must not run)", tc.name, hits.Load(), wantHits)
	}

	var report struct {
		Compatibility string `json:"compatibility"`
		Observations  struct {
			JSON struct {
				Status string `json:"status"`
				Reason string `json:"reason"`
			} `json:"json"`
			Strict struct {
				Status string `json:"status"`
				Reason string `json:"reason"`
			} `json:"strict"`
		} `json:"observations"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &report); err != nil {
		t.Fatalf("%s report = %s err=%v", tc.name, rec.Body.String(), err)
	}
	if tc.reloadReport {
		persisted := mustPersistedCapabilityReport(t, handler, profile.ID)
		if persisted.Compatibility != report.Compatibility ||
			string(persisted.Observations.JSON.Status) != report.Observations.JSON.Status ||
			persisted.Observations.JSON.Reason != report.Observations.JSON.Reason {
			t.Fatalf("%s persisted report %+v differs from returned report %+v", tc.name, persisted, report)
		}
	}

	if report.Observations.JSON.Reason != tc.wantReason {
		t.Fatalf("%s json reason = %q, want %q (report=%+v)", tc.name, report.Observations.JSON.Reason, tc.wantReason, report)
	}
	if report.Observations.Strict.Status != "unverified" {
		t.Fatalf("%s strict must remain unverified, got %+v", tc.name, report.Observations.Strict)
	}
	if report.Observations.Strict.Status == "supported" || report.Observations.JSON.Status == "supported" {
		t.Fatalf("%s must not reinterpret failure as JSON/strict success", tc.name)
	}
	if report.Observations.JSON.Reason == "unsupported_mode" || report.Observations.Strict.Reason == "unsupported_mode" {
		t.Fatalf("%s must not reinterpret failure as unsupported mode", tc.name)
	}
	for _, banned := range tc.forbidReasons {
		if report.Observations.JSON.Reason == banned {
			t.Fatalf("%s reason must not be %q", tc.name, banned)
		}
	}
	if report.Compatibility == "json-only compatible" || report.Compatibility == "strict-schema sample compatible" {
		t.Fatalf("%s must not claim compatibility pass: %q", tc.name, report.Compatibility)
	}
}

func mustPersistedCapabilityReport(t *testing.T, handler http.Handler, profileID string) *capabilities.Report {
	t.Helper()
	list := aiRequest(handler, "GET", "/ai/providers", "", aiTestOrigin, true)
	if list.Code != http.StatusOK {
		t.Fatalf("list = %d %s", list.Code, list.Body.String())
	}
	var payload struct {
		Items []aiProviderProfile `json:"items"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	for _, item := range payload.Items {
		if item.ID == profileID && item.CapabilityReport != nil {
			return item.CapabilityReport
		}
	}
	t.Fatalf("persisted capability report missing from listing for %s", profileID)
	return nil
}
