package providerhttp

import (
	"net/url"
	"testing"

	"immich-places-backend/internal/ai/providers"
)

func TestHostHeaderPreservesApprovedAuthority(t *testing.T) {
	for _, raw := range []string{
		"http://provider.example:443/v1", "https://provider.example:80/v1",
		"https://[::1]:443/v1", "http://[::1]:80/v1", "http://codex-proxy:3466/v1",
	} {
		t.Run(raw, func(t *testing.T) {
			canonical, err := providers.CanonicalBaseURL(raw)
			if err != nil {
				t.Fatal(err)
			}
			parsed, err := url.Parse(canonical.Scheme + "://" + providerAuthority(canonical))
			if err != nil || parsed.Hostname() != canonical.Host || parsed.Port() != canonical.Port {
				t.Fatalf("Host header changed the approved host/port: %q (%v)", providerAuthority(canonical), err)
			}
		})
	}
}
