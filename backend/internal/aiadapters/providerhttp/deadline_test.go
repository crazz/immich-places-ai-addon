package providerhttp

import (
	"context"
	"net"
	"testing"
	"time"

	"immich-places-backend/internal/ai/providers"
)

func TestDirectTransportEntryPointsBoundDNS(t *testing.T) {
	for _, entry := range []string{"pin", "send"} {
		t.Run(entry, func(t *testing.T) {
			called := false
			latest := time.Now().Add(121 * time.Second)
			client := New(Options{LookupIP: func(ctx context.Context, _ string) ([]net.IP, error) {
				called = true
				deadline, ok := ctx.Deadline()
				if !ok || deadline.After(latest) {
					t.Fatal("resolver must receive a bounded context")
				}
				return nil, context.DeadlineExceeded
			}})
			rule := providers.EgressRule{BaseURL: "https://provider.example:443/v1"}
			if entry == "pin" {
				_, _ = client.Pin(context.Background(), rule)
			} else {
				_, _ = client.Send(context.Background(), rule, nil)
			}
			if !called {
				t.Fatal("test did not reach DNS")
			}
		})
	}
}
