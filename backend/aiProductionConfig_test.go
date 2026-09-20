package main

import "testing"

func TestAIProductionWorkerConfigurationRejectsUnboundedSettings(t *testing.T) {
	withCleanWorkDir(t)
	t.Setenv("IMMICH_URL", "http://test:2283")
	t.Setenv("ENCRYPTION_KEY", "test-secret")
	for key, value := range map[string]string{"AI_JOB_WORKERS": "17", "AI_JOB_PER_OWNER": "3", "AI_JOB_LEASE_SECONDS": "120", "AI_JOB_HEARTBEAT_SECONDS": "0", "AI_JOB_IDLE_MS": "0"} {
		t.Run(key, func(t *testing.T) {
			t.Setenv(key, value)
			if cfg, err := loadConfig(); err == nil || cfg != nil {
				t.Fatal("invalid worker settings accepted", key)
			}
		})
	}
}
