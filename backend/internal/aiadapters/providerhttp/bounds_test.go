package providerhttp_test

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestCompleteOneBoundedProviderOperation(t *testing.T) {
	var providerHits atomic.Int64
	var seenBody []byte
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	provider := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerHits.Add(1)
		seenBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"cmpl-1","choices":[{"message":{"content":"ok"}}]}`))
	}))
	provider.Listener = listener
	provider.Start()
	t.Cleanup(provider.Close)

	port := urlPort(t, provider.URL)
	rule := mustLocalRule(t, "http://127.0.0.1:"+port+"/v1", "127.0.0.0/8")
	client := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})
	requestBody := []byte(`{"model":"vision-exact","messages":[{"role":"user","content":"locate"}]}`)
	result, err := client.Send(context.Background(), rule, requestBody)
	if err != nil {
		t.Fatalf("bounded operation failed: %v", err)
	}
	if providerHits.Load() != 1 {
		t.Fatalf("expected exactly one provider request, got %d", providerHits.Load())
	}
	if string(seenBody) != string(requestBody) {
		t.Fatalf("request body altered or model substituted: %s", seenBody)
	}
	var payload map[string]any
	if err := json.Unmarshal(result.Body, &payload); err != nil || payload["id"] != "cmpl-1" {
		t.Fatalf("bounded response missing: body=%s err=%v", result.Body, err)
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", result.StatusCode)
	}
}

func TestRejectPayloadsAndStopSlowResponses(t *testing.T) {
	t.Run("oversized request", func(t *testing.T) {
		var hits atomic.Int64
		var lookups atomic.Int64
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			hits.Add(1)
		}))
		server.Listener = listener
		server.Start()
		t.Cleanup(server.Close)

		port := urlPort(t, server.URL)
		rule := mustLocalRule(t, "http://127.0.0.1:"+port+"/v1", "127.0.0.0/8")
		client := providerhttp.New(providerhttp.Options{
			LookupIP: func(context.Context, string) ([]net.IP, error) {
				lookups.Add(1)
				return []net.IP{net.ParseIP("127.0.0.1")}, nil
			},
		})
		body := make([]byte, providers.MaxRequestBytes+1)
		result, err := client.Send(context.Background(), rule, body)
		if err == nil {
			t.Fatalf("expected limit failure, got result=%+v", result)
		}
		failure, ok := providers.AsTransportFailure(err)
		if !ok || failure.Category != providers.FailureLimit || failure.RequestID == "" {
			t.Fatalf("expected limit failure, got ok=%v %+v err=%v", ok, failure, err)
		}
		if hits.Load() != 0 {
			t.Fatalf("oversized request was transmitted: hits=%d", hits.Load())
		}
		if lookups.Load() != 0 {
			t.Fatalf("oversized request contacted DNS before limit: lookups=%d", lookups.Load())
		}
		if result.Body != nil {
			t.Fatalf("partial response reported as success: %q", result.Body)
		}
	})

	t.Run("oversized response", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"leak":"` + strings.Repeat("x", 1<<20) + `"}`))
		}))
		server.Listener = listener
		server.Start()
		t.Cleanup(server.Close)

		port := urlPort(t, server.URL)
		rule := mustLocalRule(t, "http://127.0.0.1:"+port+"/v1", "127.0.0.0/8")
		client := providerhttp.New(providerhttp.Options{
			LookupIP: func(context.Context, string) ([]net.IP, error) {
				return []net.IP{net.ParseIP("127.0.0.1")}, nil
			},
		})
		result, err := client.Send(context.Background(), rule, []byte(`{}`))
		if err == nil {
			t.Fatalf("expected limit failure, got result=%+v", result)
		}
		failure, ok := providers.AsTransportFailure(err)
		if !ok || failure.Category != providers.FailureLimit || failure.RequestID == "" {
			t.Fatalf("expected limit failure, got ok=%v %+v err=%v", ok, failure, err)
		}
		if result.Body != nil {
			t.Fatalf("partial response reported as success: len=%d", len(result.Body))
		}
		if strings.Contains(err.Error(), strings.Repeat("x", 32)) {
			t.Fatalf("failure leaked response body: %q", err.Error())
		}
	})

	t.Run("oversized headers", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("X-Pad", strings.Repeat("h", 32<<10))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
		}))
		server.Listener = listener
		server.Start()
		t.Cleanup(server.Close)

		port := urlPort(t, server.URL)
		rule := mustLocalRule(t, "http://127.0.0.1:"+port+"/v1", "127.0.0.0/8")
		client := providerhttp.New(providerhttp.Options{
			LookupIP: func(context.Context, string) ([]net.IP, error) {
				return []net.IP{net.ParseIP("127.0.0.1")}, nil
			},
		})
		result, err := client.Send(context.Background(), rule, []byte(`{}`))
		if err == nil {
			t.Fatalf("expected header limit failure, got result=%+v", result)
		}
		failure, ok := providers.AsTransportFailure(err)
		if !ok || failure.Category != providers.FailureLimit || failure.RequestID == "" {
			t.Fatalf("expected limit failure, got ok=%v %+v err=%v", ok, failure, err)
		}
		if result.Body != nil {
			t.Fatalf("partial response reported as success: %s", result.Body)
		}
	})

	t.Run("expired deadline", func(t *testing.T) {
		block := make(chan struct{})
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			<-block
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
		}))
		server.Listener = listener
		server.Start()
		t.Cleanup(func() {
			close(block)
			server.Close()
		})

		port := urlPort(t, server.URL)
		rule := mustLocalRule(t, "http://127.0.0.1:"+port+"/v1", "127.0.0.0/8")
		client := providerhttp.New(providerhttp.Options{
			LookupIP: func(context.Context, string) ([]net.IP, error) {
				return []net.IP{net.ParseIP("127.0.0.1")}, nil
			},
		})
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		result, err := client.Send(ctx, rule, []byte(`{}`))
		if err == nil {
			t.Fatalf("expected timeout failure, got result=%+v", result)
		}
		failure, ok := providers.AsTransportFailure(err)
		if !ok || failure.Category != providers.FailureTimeout || failure.RequestID == "" {
			t.Fatalf("expected timeout failure, got ok=%v %+v err=%v", ok, failure, err)
		}
		if result.Body != nil {
			t.Fatalf("partial response reported as success")
		}
	})

	t.Run("caller cancellation", func(t *testing.T) {
		started := make(chan struct{})
		block := make(chan struct{})
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			close(started)
			<-block
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
		}))
		server.Listener = listener
		server.Start()
		t.Cleanup(func() {
			close(block)
			server.Close()
		})

		port := urlPort(t, server.URL)
		rule := mustLocalRule(t, "http://127.0.0.1:"+port+"/v1", "127.0.0.0/8")
		client := providerhttp.New(providerhttp.Options{
			LookupIP: func(context.Context, string) ([]net.IP, error) {
				return []net.IP{net.ParseIP("127.0.0.1")}, nil
			},
		})
		ctx, cancel := context.WithCancel(context.Background())
		errCh := make(chan error, 1)
		go func() {
			_, sendErr := client.Send(ctx, rule, []byte(`{}`))
			errCh <- sendErr
		}()
		<-started
		cancel()
		err = <-errCh
		if err == nil {
			t.Fatalf("expected cancellation failure")
		}
		failure, ok := providers.AsTransportFailure(err)
		if !ok || failure.Category != providers.FailureCanceled || failure.RequestID == "" {
			t.Fatalf("expected canceled failure, got ok=%v %+v err=%v", ok, failure, err)
		}
	})
}
