package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestCapabilitySessionRevocationDuringDNSPreventsTransmission(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte(`{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":"color: blue\nshape: circle"}}]}`))
	}))
	t.Cleanup(server.Close)
	endpoint, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	baseURL := "http://provider.test:" + endpoint.Port() + "/v1"
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"` + baseURL + `","addressClass":"local","allowedCIDRs":["127.0.0.1/32"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	var db *Database
	transport := providerhttp.New(providerhttp.Options{LookupIP: func(ctx context.Context, _ string) ([]net.IP, error) {
		hash := sha256.Sum256([]byte("ai-session"))
		if err := db.deleteSession(ctx, hex.EncodeToString(hash[:])); err != nil {
			return nil, err
		}
		return []net.IP{net.ParseIP("127.0.0.1")}, nil
	}})
	db, handler := aiCapabilityTestHandler(t, true, policy, transport)
	profile := mustCreateProvider(t, handler, providers.Input{
		Config: providers.Config{Name: "Owned", BaseURL: baseURL, Model: "vision"}, Enabled: true,
	})
	response := aiRequest(handler, "POST", "/ai/providers/"+profile.ID+"/test", `{"expectedRevision":1}`, aiTestOrigin, true)
	if response.Code != http.StatusOK {
		t.Fatalf("accepted test response = %d %s", response.Code, response.Body.String())
	}
	if hits.Load() != 0 {
		t.Fatalf("revoked session sent %d provider request(s) after DNS; want zero", hits.Load())
	}
}
