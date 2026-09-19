package providers

import "testing"

func TestRejectMismatchedOrAmbiguousDestinations(t *testing.T) {
	policy, err := ParseEgressPolicy(`[{"baseURL":"https://provider.example:443/v1","addressClass":"public"}]`)
	if err != nil || len(policy.Rules) != 1 {
		t.Fatalf("setup policy: %+v err=%v", policy, err)
	}
	for _, tc := range []struct {
		name, baseURL string
	}{
		{"different host", "https://other.example/v1"},
		{"different scheme", "http://provider.example/v1"},
		{"different port", "https://provider.example:8443/v1"},
		{"different path", "https://provider.example/v2"},
		{"credentials", "https://user:secret@provider.example/v1"},
		{"query", "https://provider.example/v1?x=1"},
		{"fragment", "https://provider.example/v1#x"},
		{"encoded separator", "https://provider.example/v1%2F../admin"},
		{"dot segments", "https://provider.example/v1/../secret"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := MatchEgressDestination(policy, tc.baseURL); err == nil {
				t.Fatal("expected destination rejection")
			}
		})
	}
	if _, err := MatchEgressDestination(policy, "https://provider.example/v1"); err != nil {
		t.Fatalf("exact approved destination must match: %v", err)
	}
}
