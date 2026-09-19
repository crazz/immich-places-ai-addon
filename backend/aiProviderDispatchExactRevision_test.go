package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestDispatchTheExactAuthorizedRevision(t *testing.T) {
	type sink struct {
		hits int
		auth string
		body []byte
	}
	newSink := func() (*httptest.Server, *sink, string) {
		t.Helper()
		received := &sink{}
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			received.hits++
			received.auth = r.Header.Get("Authorization")
			received.body, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
		}))
		server.Listener = listener
		server.Start()
		t.Cleanup(server.Close)
		base, err := url.Parse(server.URL)
		if err != nil {
			t.Fatal(err)
		}
		return server, received, "http://127.0.0.1:" + base.Port() + "/v1"
	}

	_, oldSink, oldURL := newSink()
	_, newSinkRecv, newURL := newSink()

	policyJSON := `[{"baseURL":"` + oldURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]},{"baseURL":"` + newURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`
	policy, err := providers.ParseEgressPolicy(policyJSON)
	if err != nil {
		t.Fatal(err)
	}
	db := newTestDB(t)
	ctx := context.Background()
	first, err := db.createAIProvider(ctx, testUserID, "multi", providers.Input{
		Config:  providers.Config{Name: "Old", BaseURL: oldURL, Model: "model-v1"},
		Enabled: true,
		Secret:  ptr("secret-v1"),
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := db.updateAIProvider(ctx, testUserID, first.ID, first.Revision, providers.Input{
		Config:  providers.Config{Name: "New", BaseURL: newURL, Model: "model-v2"},
		Enabled: true,
		Secret:  ptr("secret-v2"),
	})
	if err != nil || second.Revision != 2 {
		t.Fatalf("second revision: %+v err=%v", second, err)
	}

	transport := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})
	dispatcher := newAIProviderDispatcher(db, true, policy, transport)
	_, err = dispatcher.Dispatch(ctx, providers.DispatchRequest{
		OwnerID: testUserID, ProfileID: first.ID, Revision: first.Revision,
		Body: []byte(`{"model":"model-v1"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if oldSink.hits != 1 || newSinkRecv.hits != 0 || oldSink.auth != "Bearer secret-v1" {
		t.Fatalf("revision 1 substituted: old=%+v new=%+v", oldSink, newSinkRecv)
	}
	if string(oldSink.body) != `{"model":"model-v1"}` {
		t.Fatalf("revision 1 body=%s", oldSink.body)
	}

	oldSink.hits, newSinkRecv.hits = 0, 0
	_, err = dispatcher.Dispatch(ctx, providers.DispatchRequest{
		OwnerID: testUserID, ProfileID: first.ID, Revision: second.Revision,
		Body: []byte(`{"model":"model-v2"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if newSinkRecv.hits != 1 || oldSink.hits != 0 || newSinkRecv.auth != "Bearer secret-v2" {
		t.Fatalf("revision 2 wrong destination: old=%+v new=%+v", oldSink, newSinkRecv)
	}
}
