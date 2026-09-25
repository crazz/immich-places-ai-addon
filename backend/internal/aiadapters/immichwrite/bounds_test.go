package immichwrite

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writepreview"
)

func TestMutationTimeoutAndResponseBoundsNeverResend(t *testing.T) {
	for _, kind := range []string{"timeout", "headers", "body"} {
		t.Run(kind, func(t *testing.T) {
			var calls atomic.Int32
			entered := make(chan struct{})
			release := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(out http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				close(entered)
				switch kind {
				case "timeout":
					<-release
				case "headers":
					out.Header().Set("X-Oversized", strings.Repeat("x", 32<<10))
					out.WriteHeader(200)
				case "body":
					out.Write([]byte(strings.Repeat("x", 128<<10)))
				}
			}))
			defer server.Close()
			defer close(release)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan Outcome, 1)
			go func() {
				done <- (&Transport{Endpoint: server.URL}).Send(ctx, "synthetic", "aaaaaaaa-0000-4000-8000-000000000001", writepreview.Point{Latitude: 0, Longitude: 0})
			}()
			<-entered
			if kind == "timeout" {
				cancel()
			}
			var outcome Outcome
			select {
			case outcome = <-done:
			case <-time.After(time.Second):
				t.Fatal("response was not bounded")
			}
			if calls.Load() != 1 {
				t.Fatal("hidden resend", calls.Load())
			}
			if kind != "body" && outcome.CompletionKnown {
				t.Fatal("failed I/O claimed completion", outcome)
			}
		})
	}
}
