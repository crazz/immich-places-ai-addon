package providers

import "testing"

func TestCanonicalDestinationRoundTripsAndRejectsAmbiguity(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{"http://[::1]:3466/v1", "http://[::1]:3466/v1"},
		{"https://PROVIDER.example:00443/v1.0/", "https://provider.example:443/v1.0"},
		{"http://codex-proxy:3466/v1", "http://codex-proxy:3466/v1"},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			canonical, err := CanonicalBaseURL(tc.raw)
			if err != nil || canonical.String() != tc.want {
				t.Fatalf("got %q, %v; want %q", canonical.String(), err, tc.want)
			}
			roundTrip, err := CanonicalBaseURL(canonical.String())
			if err != nil || roundTrip != canonical {
				t.Fatalf("canonical destination cannot round trip: %+v %v", roundTrip, err)
			}
		})
	}
	for _, raw := range []string{
		"https://provider.example:0/v1", "https://provider.example:65536/v1",
		"https://provider.example:/v1", "https://провайдер.example/v1",
		"http://127.1/v1", "http://2130706433/v1", "http://0177.0.0.1/v1",
		"http://0x7f000001/v1", "http://::1/v1", "https://provider.example/v1/./test",
	} {
		t.Run(raw, func(t *testing.T) {
			if _, err := CanonicalBaseURL(raw); err == nil {
				t.Fatal("ambiguous or invalid destination accepted")
			}
		})
	}
}
