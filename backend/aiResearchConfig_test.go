package main

import (
	"testing"
	"time"
)

func TestAIResearchTimeoutConfigurationIsFiniteAndDefaultsToTenMinutes(t *testing.T) {
	t.Setenv("IMMICH_URL", "http://immich.test")
	t.Setenv("ENCRYPTION_KEY", "test-key")
	for _, tc := range []struct {
		seconds string
		want    time.Duration
		invalid bool
	}{
		{"600", 10 * time.Minute, false}, {"240", 4 * time.Minute, false}, {"0", 0, true}, {"601", 0, true}, {"-1", 0, true},
	} {
		t.Setenv("AI_RESEARCH_TIMEOUT_SECONDS", tc.seconds)
		cfg, err := loadConfig()
		if tc.invalid {
			if err == nil {
				t.Fatalf("accepted %s seconds", tc.seconds)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if cfg.AIJobSettings.Policy.ResearchDuration != tc.want {
			t.Fatal("Research timeout setting ignored")
		}
	}
}
