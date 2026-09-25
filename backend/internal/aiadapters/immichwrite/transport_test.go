package immichwrite

import (
	"context"
	"encoding/json"
	"fmt"
	"immich-places-backend/internal/ai/writepreview"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestExactSingleAssetGPSOnly(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var body map[string]float64
		if r.Method != "PATCH" || r.URL.RequestURI() != "/api/assets/aaaaaaaa-0000-4000-8000-000000000001" || r.Header.Get("x-api-key") != "synthetic-key" || r.Header.Get("Idempotency-Key") != "" || json.NewDecoder(r.Body).Decode(&body) != nil || len(body) != 2 || body["latitude"] != 0 || body["longitude"] != 12 {
			t.Error("unexpected mutation", r.Method, r.URL, body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer server.Close()
	out := (&Transport{Endpoint: server.URL}).Send(context.Background(), "synthetic-key", "aaaaaaaa-0000-4000-8000-000000000001", writepreview.Point{Latitude: 0, Longitude: 12})
	if calls != 1 || !out.CompletionKnown || out.Code != "response_received" {
		t.Fatal(calls, out)
	}
}

func TestMutationNeverRetriesRedirectsDropsOrFailures(t *testing.T) {
	for _, status := range []int{301, 307, 429, 500, 503, 0} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if status == 0 {
					conn, _, err := w.(http.Hijacker).Hijack()
					if err != nil {
						t.Error(err)
						return
					}
					conn.Close()
					return
				}
				w.Header().Set("Location", "/alternate")
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(status)
			}))
			defer server.Close()
			_ = (&Transport{Endpoint: server.URL}).Send(context.Background(), "key", "aaaaaaaa-0000-4000-8000-000000000001", writepreview.Point{Latitude: 1, Longitude: 2})
			if calls.Load() != 1 {
				t.Fatal("hidden sends", calls.Load())
			}
		})
	}
}
