package providers_test

import (
	"strings"
	"testing"

	"immich-places-backend/internal/ai/providers"
)

func TestFingerprintPolicyIsOpaqueAndStable(t *testing.T) {
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"http://codex-proxy:3466/v1","addressClass":"local","allowedCIDRs":["192.168.144.0/20"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	first := providers.FingerprintPolicy(policy)
	second := providers.FingerprintPolicy(policy)
	if first == "" || first != second {
		t.Fatalf("fingerprint unstable: %q vs %q", first, second)
	}
	if len(first) != 64 {
		t.Fatalf("expected sha256 hex fingerprint, got %q", first)
	}
	for _, leak := range []string{"allowedCIDRs", "192.168.144", "codex-proxy", "{", "["} {
		if strings.Contains(first, leak) {
			t.Fatalf("fingerprint leaked %q: %s", leak, first)
		}
	}
}
