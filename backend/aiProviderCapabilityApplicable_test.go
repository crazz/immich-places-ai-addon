package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestCompletedUnverifiedReportIsNotApplicableOnReload(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"cmpl-1","model":"gpt-5.6-sol","choices":[{"message":{"role":"assistant","content":"not a fixture answer"},"finish_reason":"stop"}]}`))
	}))
	proxy.Listener = listener
	proxy.Start()
	t.Cleanup(proxy.Close)
	base, err := url.Parse(proxy.URL)
	if err != nil {
		t.Fatal(err)
	}
	approvedURL := "http://codex-proxy:" + base.Port() + "/v1"
	policyJSON := `[{"baseURL":"` + approvedURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`
	policy, err := providers.ParseEgressPolicy(policyJSON)
	if err != nil {
		t.Fatal(err)
	}
	transport := providerhttp.New(providerhttp.Options{
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

	tested := aiRequest(handler, "POST", "/ai/providers/"+profile.ID+"/test", mustJSON(t, map[string]any{"expectedRevision": 1}), aiTestOrigin, true)
	var postReport struct {
		Applicable   bool   `json:"applicable"`
		Lifecycle    string `json:"lifecycle"`
		Fingerprint  string `json:"policyFingerprint"`
		Observations struct {
			Image struct {
				Status string `json:"status"`
			} `json:"image"`
		} `json:"observations"`
	}
	if err := json.Unmarshal(tested.Body.Bytes(), &postReport); err != nil || tested.Code != 200 {
		t.Fatalf("post test = %s err=%v", tested.Body.String(), err)
	}
	if postReport.Lifecycle != "completed" || postReport.Observations.Image.Status != "unverified" {
		t.Fatalf("expected completed unverified observation: %+v", postReport)
	}
	if postReport.Applicable {
		t.Fatal("POST must not mark unverified evidence applicable")
	}
	if postReport.Fingerprint == "" || strings.Contains(postReport.Fingerprint, "allowedCIDRs") || strings.Contains(postReport.Fingerprint, "127.0.0.0") || strings.Contains(postReport.Fingerprint, approvedURL) {
		t.Fatalf("policy fingerprint must be opaque, got %q", postReport.Fingerprint)
	}

	listed := aiRequest(handler, "GET", "/ai/providers", "", "", true)
	var list struct {
		Items []struct {
			CapabilityReport *struct {
				Applicable   bool   `json:"applicable"`
				Fingerprint  string `json:"policyFingerprint"`
				Observations struct {
					Image struct {
						Status string `json:"status"`
					} `json:"image"`
				} `json:"observations"`
			} `json:"capabilityReport"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listed.Body.Bytes(), &list); err != nil || listed.Code != 200 || list.Items[0].CapabilityReport == nil {
		t.Fatalf("reload list = %s err=%v", listed.Body.String(), err)
	}
	reloaded := list.Items[0].CapabilityReport
	if reloaded.Observations.Image.Status != "unverified" {
		t.Fatalf("reloaded observation = %+v", reloaded.Observations)
	}
	if reloaded.Applicable {
		t.Fatal("reload must not mark completed unverified evidence applicable")
	}
	if reloaded.Fingerprint != postReport.Fingerprint {
		t.Fatalf("fingerprint changed across reload: post=%q list=%q", postReport.Fingerprint, reloaded.Fingerprint)
	}
	if strings.Contains(listed.Body.String(), "allowedCIDRs") || strings.Contains(listed.Body.String(), "127.0.0.0/8") {
		t.Fatal("raw egress policy leaked in list response")
	}
}

func TestDisablementInvalidatesProofWithoutRetest(t *testing.T) {
	hits := &atomic.Int32{}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capabilityFixtureResponse(w, r, hits)
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
	transport := providerhttp.New(providerhttp.Options{
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
	success := aiRequest(handler, "POST", "/ai/providers/"+profile.ID+"/test", mustJSON(t, map[string]any{"expectedRevision": 1}), aiTestOrigin, true)
	var before struct {
		Applicable bool `json:"applicable"`
	}
	if err := json.Unmarshal(success.Body.Bytes(), &before); err != nil || success.Code != 200 || !before.Applicable {
		t.Fatalf("successful proof = %s err=%v", success.Body.String(), err)
	}
	afterSuccess := hits.Load()

	disabled := aiRequest(handler, "PUT", "/ai/providers/"+profile.ID, mustJSON(t, aiProviderRequest{
		Input: providers.Input{
			Config:  providers.Config{Name: "Owned", BaseURL: approvedURL, Model: "gpt-5.6-sol"},
			Enabled: false,
		},
		ExpectedRevision: 1,
	}), aiTestOrigin, true)
	if disabled.Code != 200 {
		t.Fatalf("disable = %s", disabled.Body.String())
	}
	if hits.Load() != afterSuccess {
		t.Fatal("disablement retested automatically")
	}
	listed := aiRequest(handler, "GET", "/ai/providers", "", "", true)
	var list struct {
		Items []struct {
			Enabled          bool `json:"enabled"`
			CapabilityReport *struct {
				Applicable bool `json:"applicable"`
			} `json:"capabilityReport"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listed.Body.Bytes(), &list); err != nil || listed.Code != 200 {
		t.Fatalf("list after disable = %s err=%v", listed.Body.String(), err)
	}
	if list.Items[0].Enabled {
		t.Fatal("profile should be disabled")
	}
	// Disablement creates a new revision; current settings need their own test.
	if list.Items[0].CapabilityReport != nil && list.Items[0].CapabilityReport.Applicable {
		t.Fatal("disabled settings must not present prior proof as current usable proof")
	}
	if hits.Load() != afterSuccess {
		t.Fatal("loading disabled settings retested automatically")
	}
}

func TestCapabilityStartupStorageFailureMakesZeroProviderRequests(t *testing.T) {
	hits := &atomic.Int32{}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capabilityFixtureResponse(w, r, hits)
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
	transport := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})
	db, handler := aiCapabilityTestHandler(t, true, policy, transport)
	secret := "secret"
	profile := mustCreateProvider(t, handler, providers.Input{
		Config:  providers.Config{Name: "Owned", BaseURL: approvedURL, Model: "gpt-5.6-sol"},
		Enabled: true,
		Secret:  &secret,
	})
	if _, err := db.db.Exec(`DROP TABLE ai_provider_capability_checks`); err != nil {
		t.Fatal(err)
	}
	failed := aiRequest(handler, "POST", "/ai/providers/"+profile.ID+"/test", mustJSON(t, map[string]any{"expectedRevision": 1}), aiTestOrigin, true)
	if failed.Code != http.StatusInternalServerError {
		t.Fatalf("startup storage failure status = %d %s", failed.Code, failed.Body.String())
	}
	var failure struct{ Code string }
	if err := json.Unmarshal(failed.Body.Bytes(), &failure); err != nil || failure.Code != "STORAGE_ERROR" {
		t.Fatalf("startup storage failure body = %s", failed.Body.String())
	}
	if hits.Load() != 0 {
		t.Fatalf("startup storage failure made %d provider requests", hits.Load())
	}
}

func TestCapabilityCompletionStorageFailureReportsNoDurableEvidence(t *testing.T) {
	var hits atomic.Int32
	release := make(chan struct{})
	started := make(chan struct{})
	var startOnce sync.Once
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n == 1 {
			startOnce.Do(func() { close(started) })
			<-release
		}
		capabilityFixtureResponse(w, r, nil)
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
	transport := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})
	db, handler := aiCapabilityTestHandler(t, true, policy, transport)
	secret := "secret"
	profile := mustCreateProvider(t, handler, providers.Input{
		Config:  providers.Config{Name: "Owned", BaseURL: approvedURL, Model: "gpt-5.6-sol"},
		Enabled: true,
		Secret:  &secret,
	})

	var result *httptest.ResponseRecorder
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		result = aiRequest(handler, "POST", "/ai/providers/"+profile.ID+"/test", mustJSON(t, map[string]any{"expectedRevision": 1}), aiTestOrigin, true)
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("first probe did not start")
	}
	if _, err := db.db.Exec(`DROP TABLE ai_provider_capability_checks`); err != nil {
		t.Fatal(err)
	}
	close(release)
	wg.Wait()
	if result.Code != http.StatusInternalServerError {
		t.Fatalf("completion storage failure status = %d %s", result.Code, result.Body.String())
	}
	var failure struct{ Code string }
	if err := json.Unmarshal(result.Body.Bytes(), &failure); err != nil || failure.Code != "STORAGE_ERROR" {
		t.Fatalf("completion storage failure body = %s", result.Body.String())
	}
	if hits.Load() > 3 {
		t.Fatalf("completion storage failure replayed provider requests: hits=%d", hits.Load())
	}
	// Table gone: no durable successful evidence remains.
	if strings.Contains(result.Body.String(), `"applicable":true`) || strings.Contains(result.Body.String(), `"lifecycle":"completed"`) {
		t.Fatal("storage failure must not report durable successful test")
	}
}
