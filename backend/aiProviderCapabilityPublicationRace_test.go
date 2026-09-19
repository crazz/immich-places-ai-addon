package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestCapabilityEditDuringFinalResponseCannotPublishCurrentProof(t *testing.T) {
	var db *Database
	var profile aiProviderProfile
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := hits.Add(1)
		if n == 3 {
			_, err := db.updateAIProvider(context.Background(), testUserID, profile.ID, profile.Revision, providers.Input{
				Config: profile.Config, Enabled: true,
			})
			if err != nil {
				t.Errorf("edit during final response: %v", err)
				http.Error(w, "fixture edit failed", http.StatusInternalServerError)
				return
			}
		}
		content := "color: blue\nshape: circle"
		if n > 1 {
			content = `{"color":"blue","shape":"circle"}`
		}
		encoded, _ := json.Marshal(content)
		fmt.Fprintf(w, `{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":%s}}]}`, encoded)
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
	transport := providerhttp.New(providerhttp.Options{LookupIP: func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("127.0.0.1")}, nil
	}})
	db, handler := aiCapabilityTestHandler(t, true, policy, transport)
	profile = mustCreateProvider(t, handler, providers.Input{
		Config: providers.Config{Name: "Owned", BaseURL: baseURL, Model: "vision"}, Enabled: true,
	})
	response := aiRequest(handler, "POST", "/ai/providers/"+profile.ID+"/test", `{"expectedRevision":1}`, aiTestOrigin, true)
	var report capabilities.Report
	if err := json.Unmarshal(response.Body.Bytes(), &report); err != nil || response.Code != http.StatusOK || hits.Load() != 3 {
		t.Fatalf("test response = %d %s; hits=%d err=%v", response.Code, response.Body.String(), hits.Load(), err)
	}
	if report.Applicable {
		t.Fatal("old revision published applicable proof after edit during final provider response")
	}
}
