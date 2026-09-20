package main

import (
	"strings"
	"testing"
)

func TestAIExecutionConfigurationRejectsInvalidAttestationsEvenWhenDisabled(t *testing.T) {
	withCleanWorkDir(t)
	t.Setenv("IMMICH_URL", "http://test:2283")
	t.Setenv("ENCRYPTION_KEY", "test-secret")
	t.Setenv("AI_ENABLED", "false")
	t.Setenv("AI_EXECUTION_POLICIES", `[{"evidenceRef":"private-attestation"}]`)
	cfg, err := loadConfig()
	if err == nil || cfg != nil {
		t.Fatal("invalid execution policy configuration accepted")
	}
	if strings.Contains(err.Error(), "private-attestation") {
		t.Fatal("private policy configuration leaked")
	}
	t.Setenv("AI_EXECUTION_POLICIES", "")
	if cfg, err = loadConfig(); err != nil || cfg == nil || cfg.AIEnabled {
		t.Fatal("empty policy defaults enabled work or broke disabled startup", err)
	}
}
