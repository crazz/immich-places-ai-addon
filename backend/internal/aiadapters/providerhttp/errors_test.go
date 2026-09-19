package providerhttp_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestErrorsDoNotCauseHiddenRetriesOrLeaks(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		secret string
	}{
		{"authentication", http.StatusUnauthorized, `{"error":{"message":"sk-secret-key-leaked","code":"invalid_api_key"}}`, "sk-secret-key-leaked"},
		{"rate limit", http.StatusTooManyRequests, `{"error":{"message":"slow down","code":"rate_limit_exceeded"}}`, "slow down"},
		{"server error", http.StatusInternalServerError, `{"error":{"message":"internal boom","type":"server_error"}}`, "internal boom"},
		{"unsafe body", http.StatusBadRequest, `{"error":{"message":"Bearer sk-live-abcdef remaining"}}`, "sk-live-abcdef"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var hits atomic.Int64
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				hits.Add(1)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
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
			result, err := client.Send(context.Background(), rule, []byte(`{"model":"vision-exact"}`))
			if err == nil {
				t.Fatalf("expected upstream failure, got result=%+v", result)
			}
			if hits.Load() != 1 {
				t.Fatalf("expected one request without retry, got %d", hits.Load())
			}
			if result.Body != nil {
				t.Fatalf("partial upstream body reported as success: %s", result.Body)
			}
			if strings.Contains(err.Error(), tc.secret) || strings.Contains(err.Error(), "Bearer") {
				t.Fatalf("credential-bearing diagnostic emitted: %q", err.Error())
			}
			failure, ok := providers.AsTransportFailure(err)
			if !ok {
				t.Fatalf("expected typed transport failure, got %T %v", err, err)
			}
			if failure.Category != providers.FailureUpstream || failure.RequestID == "" {
				t.Fatalf("unstable failure shape: %+v", failure)
			}
			if failure.Status != tc.status {
				t.Fatalf("status=%d want=%d", failure.Status, tc.status)
			}
		})
	}

	t.Run("lost response", func(t *testing.T) {
		var hits atomic.Int64
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			hits.Add(1)
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("hijack unsupported")
			}
			conn, _, hijackErr := hj.Hijack()
			if hijackErr != nil {
				t.Fatal(hijackErr)
			}
			_ = conn.Close()
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
		result, err := client.Send(context.Background(), rule, []byte(`{"Authorization":"Bearer must-not-retry"}`))
		if err == nil {
			t.Fatalf("expected lost-response failure, got result=%+v", result)
		}
		if hits.Load() != 1 {
			t.Fatalf("lost response triggered retries: hits=%d", hits.Load())
		}
		if strings.Contains(err.Error(), "must-not-retry") || strings.Contains(err.Error(), "Bearer") {
			t.Fatalf("credential-bearing diagnostic emitted: %q", err.Error())
		}
		failure, ok := providers.AsTransportFailure(err)
		if !ok || failure.Category != providers.FailureConnection || failure.RequestID == "" {
			t.Fatalf("expected connection failure with request id, got ok=%v %+v", ok, failure)
		}
	})
}
