package providerhttp_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestRetainAllowlistedUpstreamErrorCode(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"sk-secret-should-not-leak","code":"invalid_api_key"}}`))
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
	_, err = client.Send(context.Background(), rule, []byte(`{}`))
	failure, ok := providers.AsTransportFailure(err)
	if !ok || failure.Category != providers.FailureUpstream || failure.Status != http.StatusUnauthorized {
		t.Fatalf("expected upstream failure, got ok=%v %+v err=%v", ok, failure, err)
	}
	if failure.Code != "invalid_api_key" {
		t.Fatalf("allowlisted code=%q", failure.Code)
	}
	if failure.Error() != "provider UPSTREAM failure" {
		t.Fatalf("unsafe error text: %q", failure.Error())
	}
}
