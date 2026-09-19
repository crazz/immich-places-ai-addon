package providerhttp_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestCapabilityResponseCeilingRejectsBodyBelowGeneralDefault(t *testing.T) {
	rule := oversizedCapabilityResponseServer(t)
	client := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
		MaxResponseBytes: providerhttp.MaxCapabilityResponseBytes,
	})

	result, err := client.Send(context.Background(), rule, []byte(`{}`))
	if err == nil {
		t.Fatalf("expected capability response limit failure, got body length %d", len(result.Body))
	}
	failure, ok := providers.AsTransportFailure(err)
	if !ok || failure.Category != providers.FailureLimit || failure.RequestID == "" {
		t.Fatalf("expected limit failure, got ok=%v %+v err=%v", ok, failure, err)
	}
	if result.Body != nil {
		t.Fatalf("oversized capability response reported as success: len=%d", len(result.Body))
	}
}

func TestGeneralResponseCeilingStillAcceptsCapabilitySizedBody(t *testing.T) {
	rule := oversizedCapabilityResponseServer(t)
	client := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})

	result, err := client.Send(context.Background(), rule, []byte(`{}`))
	if err != nil {
		t.Fatalf("general transport default must still accept this body: %v", err)
	}
	if len(result.Body) <= providerhttp.MaxCapabilityResponseBytes {
		t.Fatalf("body length %d does not exceed the capability ceiling", len(result.Body))
	}
}

func oversizedCapabilityResponseServer(t *testing.T) providers.EgressRule {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pad":"` + strings.Repeat("x", providerhttp.MaxCapabilityResponseBytes) + `"}`))
	}))
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)
	return mustLocalRule(t, "http://127.0.0.1:"+urlPort(t, server.URL)+"/v1", "127.0.0.0/8")
}
