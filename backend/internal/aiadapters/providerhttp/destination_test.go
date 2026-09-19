package providerhttp_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestResolveApprovedDockerHostname(t *testing.T) {
	var dialed atomic.Value
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dialed.Store(r.Context().Value(http.LocalAddrContextKey))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)

	port := urlPort(t, server.URL)
	rule := mustLocalRule(t, "http://codex-proxy:"+port+"/v1", "127.0.0.0/8")
	client := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})
	result, err := client.Send(context.Background(), rule, []byte(`{}`))
	if err != nil || result.StatusCode != http.StatusOK {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	local, _ := dialed.Load().(net.Addr)
	if local == nil || local.String() != listener.Addr().String() {
		t.Fatalf("expected dial to validated listener %s, got %v", listener.Addr(), local)
	}
}

func TestRejectChangedMixedOrAbsentDNSAnswers(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits++ }))
	t.Cleanup(server.Close)
	port := urlPort(t, server.URL)
	rule := mustLocalRule(t, "http://codex-proxy:"+port+"/v1", "127.0.0.0/8")

	for _, tc := range []struct {
		name string
		ips  []net.IP
	}{
		{"unapproved", []net.IP{net.ParseIP("8.8.8.8")}},
		{"mixed", []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("8.8.8.8")}},
		{"absent", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lookups := 0
			client := providerhttp.New(providerhttp.Options{
				LookupIP: func(context.Context, string) ([]net.IP, error) {
					lookups++
					return tc.ips, nil
				},
			})
			_, err := client.Send(context.Background(), rule, []byte(`{"secret":"no"}`))
			if err == nil || hits != 0 {
				t.Fatalf("expected rejection without HTTP: hits=%d err=%v", hits, err)
			}
			if lookups != 1 {
				t.Fatalf("expected one controlled lookup, got %d", lookups)
			}
			failure, ok := providers.AsTransportFailure(err)
			if !ok || failure.Category != providers.FailureConnection || failure.RequestID == "" {
				t.Fatalf("expected typed connection failure, got ok=%v %+v err=%v", ok, failure, err)
			}
		})
	}
}

func TestLocalExceptionCannotAuthorizeMetadataAccess(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits++ }))
	t.Cleanup(server.Close)
	port := urlPort(t, server.URL)
	baseRule := mustLocalRule(t, "http://codex-proxy:"+port+"/v1", "127.0.0.0/8")
	for _, ip := range []string{"0.0.0.0", "224.0.0.1", "169.254.1.1", "169.254.169.254", "100.100.100.200", "::ffff:169.254.169.254"} {
		t.Run(ip, func(t *testing.T) {
			client := providerhttp.New(providerhttp.Options{
				LookupIP: func(context.Context, string) ([]net.IP, error) {
					return []net.IP{net.ParseIP(ip)}, nil
				},
			})
			_, err := client.Send(context.Background(), baseRule, []byte(`{"Authorization":"secret"}`))
			if err == nil || hits != 0 {
				t.Fatalf("metadata path must not transmit: hits=%d err=%v", hits, err)
			}
			failure, ok := providers.AsTransportFailure(err)
			if !ok || failure.Category != providers.FailureConnection || failure.RequestID == "" {
				t.Fatalf("expected typed connection failure, got ok=%v %+v err=%v", ok, failure, err)
			}
		})
	}
}

func mustLocalRule(t *testing.T, baseURL, cidr string) providers.EgressRule {
	t.Helper()
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"` + baseURL + `","addressClass":"local","allowedCIDRs":["` + cidr + `"]}]`)
	if err != nil || len(policy.Rules) != 1 {
		t.Fatalf("rule setup: %+v err=%v", policy, err)
	}
	return policy.Rules[0]
}

func urlPort(t *testing.T, raw string) string {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return u.Port()
}
