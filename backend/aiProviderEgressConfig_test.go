package main

import (
	"strings"
	"testing"
)

func TestInvalidAdministratorPolicyFailsClosed(t *testing.T) {
	withCleanWorkDir(t)
	t.Setenv("IMMICH_URL", "http://test:2283")
	t.Setenv("ENCRYPTION_KEY", "test-secret")
	t.Setenv("AI_ENABLED", "true")
	t.Setenv("AI_PUBLIC_ORIGIN", "https://places.example")

	for _, tc := range []struct {
		name, policy string
	}{
		{"malformed JSON", `{`},
		{"wildcard host", `[{"baseURL":"https://*.example.com/v1","addressClass":"public"}]`},
		{"unbounded local CIDR", `[{"baseURL":"http://codex-proxy:3466/v1","addressClass":"local","allowedCIDRs":["0.0.0.0/0"]}]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("AI_PROVIDER_EGRESS_POLICY", tc.policy)
			_, err := loadConfig()
			if err == nil {
				t.Fatal("expected invalid policy to reject startup")
			}
			msg := err.Error()
			if !strings.Contains(msg, "AI_PROVIDER_EGRESS_POLICY") {
				t.Fatalf("error must name the setting: %v", err)
			}
			if strings.Contains(msg, "0.0.0.0/0") || strings.Contains(msg, "*.example") {
				t.Fatalf("error leaked policy detail: %v", err)
			}
		})
	}

	t.Setenv("AI_PROVIDER_EGRESS_POLICY", "[]")
	cfg, err := loadConfig()
	if err != nil || cfg == nil || len(cfg.AIProviderEgressPolicy.Rules) != 0 {
		t.Fatalf("empty valid policy must load for offline profile management: cfg=%v err=%v", cfg, err)
	}
}
