package providerhttp_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestStopEveryRedirect(t *testing.T) {
	var targetHits atomic.Int64
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		targetHits.Add(1)
	}))
	t.Cleanup(target.Close)

	for _, tc := range []struct {
		name     string
		location func(approvedURL string) string
	}{
		{"same-origin", func(approvedURL string) string { return approvedURL + "/other" }},
		{"other-origin", func(string) string { return target.URL + "/elsewhere" }},
		{"internal-service", func(string) string { return "http://127.0.0.1:9/internal" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			targetHits.Store(0)
			var approvedHits atomic.Int64
			var seenAuth string
			var seenBody []byte
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			var approvedURL string
			approved := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				approvedHits.Add(1)
				seenAuth = r.Header.Get("Authorization")
				seenBody, _ = io.ReadAll(r.Body)
				w.Header().Set("Location", tc.location(approvedURL))
				w.WriteHeader(http.StatusTemporaryRedirect)
			}))
			approved.Listener = listener
			approved.Start()
			t.Cleanup(approved.Close)
			approvedURL = approved.URL

			port := urlPort(t, approved.URL)
			rule := mustLocalRule(t, "http://127.0.0.1:"+port+"/v1", "127.0.0.0/8")
			client := providerhttp.New(providerhttp.Options{
				LookupIP: func(context.Context, string) ([]net.IP, error) {
					return []net.IP{net.ParseIP("127.0.0.1")}, nil
				},
			})
			dest, err := client.Pin(context.Background(), rule)
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.Transmit(context.Background(), dest, []byte(`{"secret":"payload"}`), "Bearer must-not-replay")
			if err == nil || !errors.Is(err, providers.ErrRedirect) {
				t.Fatalf("expected redirect failure, got %v", err)
			}
			if approvedHits.Load() != 1 {
				t.Fatalf("approved origin hits=%d", approvedHits.Load())
			}
			if targetHits.Load() != 0 {
				t.Fatalf("redirect target received %d follow-up requests", targetHits.Load())
			}
			if seenAuth != "Bearer must-not-replay" || string(seenBody) != `{"secret":"payload"}` {
				t.Fatalf("first request not observed: auth=%q body=%s", seenAuth, seenBody)
			}
		})
	}
}
