package main

import "testing"

func TestUnusedEgressPolicyDoesNotBreakAIDisabledStartup(t *testing.T) {
	withCleanWorkDir(t)
	t.Setenv("IMMICH_URL", "http://test:2283")
	t.Setenv("ENCRYPTION_KEY", "test-secret")
	t.Setenv("AI_ENABLED", "false")
	t.Setenv("AI_PROVIDER_EGRESS_POLICY", `{`)
	cfg, err := loadConfig()
	if err != nil || cfg.AIEnabled {
		t.Fatalf("AI-disabled startup must ignore unused policy: cfg=%v err=%v", cfg, err)
	}
}
