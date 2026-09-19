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

func TestRejectUnsupportedResponseEncoding(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-really-gzip"))
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
		t.Fatalf("expected encoding failure, got result=%+v", result)
	}
	failure, ok := providers.AsTransportFailure(err)
	if !ok || failure.Category != providers.FailureEncoding || failure.RequestID == "" {
		t.Fatalf("expected encoding failure, got ok=%v %+v err=%v", ok, failure, err)
	}
	if result.Body != nil {
		t.Fatalf("partial encoded body reported as success")
	}
}
