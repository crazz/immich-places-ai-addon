package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestInterruptResidualRunningCapabilityChecksOnStartup(t *testing.T) {
	db := newTestDB(t)
	profile, err := db.createAIProvider(context.Background(), testUserID, "profile", providers.Input{
		Config:  providers.Config{Name: "Owned", BaseURL: "https://provider.example/v1", Model: "vision"},
		Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now().UTC().Add(-time.Hour)
	deadline := started.Add(2 * time.Minute)
	admission, err := db.admitAIProviderCapability(context.Background(), testUserID, profile.ID, profile.Revision, "policy", started, deadline)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.interruptRunningAIProviderCapabilities(context.Background()); err != nil {
		t.Fatal(err)
	}
	loaded, err := db.loadAIProviderCapability(context.Background(), testUserID, profile.ID, profile.Revision, capabilities.ApplicabilityContext{})
	if err != nil || loaded == nil || loaded.Lifecycle != "interrupted" || loaded.AttemptID != admission.AttemptID {
		t.Fatalf("residual running attempt not interrupted: %+v err=%v", loaded, err)
	}
	var count int
	if err := db.db.QueryRow(`SELECT count(*) FROM ai_provider_capability_checks WHERE lifecycle = 'running'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("running rows remain: %d err=%v", count, err)
	}
}

func TestExpireRunningCapabilityOnReadWithoutReplay(t *testing.T) {
	db := newTestDB(t)
	profile, err := db.createAIProvider(context.Background(), testUserID, "profile", providers.Input{
		Config:  providers.Config{Name: "Owned", BaseURL: "https://provider.example/v1", Model: "vision"},
		Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now().UTC().Add(-time.Hour)
	deadline := started.Add(2 * time.Minute) // already past
	admission, err := db.admitAIProviderCapability(context.Background(), testUserID, profile.ID, profile.Revision, "policy", started, deadline)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := db.loadAIProviderCapability(context.Background(), testUserID, profile.ID, profile.Revision, capabilities.ApplicabilityContext{
		AIEnabled: true, ProfileEnabled: true, ActiveRevision: profile.Revision,
		CurrentProtocolVersion: capabilities.ProtocolVersion, CurrentPolicyFingerprint: "policy",
	})
	if err != nil || loaded == nil || loaded.Lifecycle != "interrupted" || loaded.AttemptID != admission.AttemptID {
		t.Fatalf("expired running attempt not interrupted on read: %+v err=%v", loaded, err)
	}
	if loaded.Applicable {
		t.Fatal("interrupted evidence must not claim success")
	}
	reloaded, err := db.loadAIProviderCapability(context.Background(), testUserID, profile.ID, profile.Revision, capabilities.ApplicabilityContext{
		AIEnabled: true, ProfileEnabled: true, ActiveRevision: profile.Revision,
		CurrentProtocolVersion: capabilities.ProtocolVersion, CurrentPolicyFingerprint: "policy",
	})
	if err != nil || reloaded == nil || reloaded.Lifecycle != "interrupted" || reloaded.AttemptID != admission.AttemptID {
		t.Fatalf("interrupted record unstable across reads: %+v err=%v", reloaded, err)
	}
}

func TestProfileEditBetweenProbesStopsSubsequentDispatch(t *testing.T) {
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
		body, _ := io.ReadAll(r.Body)
		content := "color: blue\nshape: circle"
		if strings.Contains(string(body), `"json_object"`) || strings.Contains(string(body), `"json_schema"`) {
			content = `{"color":"blue","shape":"circle"}`
		}
		contentJSON, _ := json.Marshal(content)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"cmpl-1","model":"gpt-5.6-sol","choices":[{"message":{"role":"assistant","content":` + string(contentJSON) + `},"finish_reason":"stop"}]}`))
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
	edited := aiRequest(handler, "PUT", "/ai/providers/"+profile.ID, mustJSON(t, aiProviderRequest{
		Input: providers.Input{
			Config:  providers.Config{Name: "Owned", BaseURL: approvedURL, Model: "gpt-5.6-sol"},
			Enabled: true,
		},
		ExpectedRevision: 1,
	}), aiTestOrigin, true)
	if edited.Code != 200 {
		t.Fatalf("edit during test = %s", edited.Body.String())
	}
	close(release)
	wg.Wait()
	if hits.Load() != 1 {
		t.Fatalf("provider hits = %d, want 1 after revision change", hits.Load())
	}
	var report capabilities.Report
	if err := json.Unmarshal(result.Body.Bytes(), &report); err != nil || result.Code != 200 {
		t.Fatalf("test response = %d %s err=%v", result.Code, result.Body.String(), err)
	}
	if !report.InputMayBeConsumed {
		t.Fatal("already transmitted input must be described as potentially consumed")
	}
	if report.Applicable {
		t.Fatal("edited revision must not publish applicable proof")
	}
	listed := aiRequest(handler, "GET", "/ai/providers", "", "", true)
	var list struct {
		Items []struct {
			Revision         int              `json:"revision"`
			CapabilityReport *json.RawMessage `json:"capabilityReport"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listed.Body.Bytes(), &list); err != nil || list.Items[0].Revision != 2 || list.Items[0].CapabilityReport != nil {
		t.Fatalf("current settings must need their own completed test: %s", listed.Body.String())
	}
}

func TestSessionRevocationBetweenProbesStopsSubsequentDispatch(t *testing.T) {
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
		body, _ := io.ReadAll(r.Body)
		content := "color: blue\nshape: circle"
		if strings.Contains(string(body), `"json_object"`) || strings.Contains(string(body), `"json_schema"`) {
			content = `{"color":"blue","shape":"circle"}`
		}
		contentJSON, _ := json.Marshal(content)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"cmpl-1","model":"gpt-5.6-sol","choices":[{"message":{"role":"assistant","content":` + string(contentJSON) + `},"finish_reason":"stop"}]}`))
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
	hash := sha256.Sum256([]byte("ai-session"))
	if err := db.deleteSession(context.Background(), hex.EncodeToString(hash[:])); err != nil {
		t.Fatal(err)
	}
	close(release)
	wg.Wait()
	if hits.Load() != 1 {
		t.Fatalf("provider hits = %d, want 1 after session revocation", hits.Load())
	}
	var report capabilities.Report
	if err := json.Unmarshal(result.Body.Bytes(), &report); err != nil || result.Code != 200 {
		t.Fatalf("test response = %d %s err=%v", result.Code, result.Body.String(), err)
	}
	if !report.InputMayBeConsumed {
		t.Fatal("already transmitted input must be described as potentially consumed")
	}
	if report.Applicable || report.Lifecycle == "completed" && report.Observations.JSON.Status == "supported" {
		t.Fatalf("revoked session must not publish continued proof: %+v", report)
	}
}
