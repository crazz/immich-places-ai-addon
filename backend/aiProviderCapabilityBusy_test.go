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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestRejectConcurrentCapabilityTestStarts(t *testing.T) {
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

	var first *httptest.ResponseRecorder
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		first = aiRequest(handler, "POST", "/ai/providers/"+profile.ID+"/test", mustJSON(t, map[string]any{"expectedRevision": 1}), aiTestOrigin, true)
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("first capability test did not start")
	}
	busy := aiRequest(handler, "POST", "/ai/providers/"+profile.ID+"/test", mustJSON(t, map[string]any{"expectedRevision": 1}), aiTestOrigin, true)
	close(release)
	wg.Wait()
	if busy.Code != http.StatusConflict {
		t.Fatalf("busy status = %d body = %s", busy.Code, busy.Body.String())
	}
	var failure struct{ Code string }
	if err := json.Unmarshal(busy.Body.Bytes(), &failure); err != nil || failure.Code != "PROVIDER_BUSY" {
		t.Fatalf("busy error = %s", busy.Body.String())
	}
	if first.Code != http.StatusOK {
		t.Fatalf("first test = %d %s", first.Code, first.Body.String())
	}
	if hits.Load() != 3 {
		t.Fatalf("provider hits = %d, want 3 for the accepted test only", hits.Load())
	}
}
