package main

import "testing"

func TestAISelectionConfigurationBounds(t *testing.T) {
	withCleanWorkDir(t)
	t.Setenv("IMMICH_URL", "https://immich.example/api")
	t.Setenv("ENCRYPTION_KEY", "test-key")
	t.Setenv("AI_ENABLED", "false")
	for _, tc := range []struct{ key, value string }{
		{"AI_SELECTION_MAX_ASSETS", "0"}, {"AI_SELECTION_MAX_ASSETS", "5001"},
		{"AI_SELECTION_TTL_SECONDS", "59"}, {"AI_SELECTION_TTL_SECONDS", "3601"},
		{"AI_INSTANCE_EPOCH", " "},
	} {
		t.Run(tc.key+tc.value, func(t *testing.T) {
			t.Setenv(tc.key, tc.value)
			if _, err := loadConfig(); err == nil {
				t.Fatalf("invalid %s=%s accepted", tc.key, tc.value)
			}
		})
	}
}
