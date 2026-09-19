package main

import "testing"

func TestAIConfigurationDefaultsAndOrigin(t *testing.T) {
	withCleanWorkDir(t)
	t.Setenv("IMMICH_URL", "http://test:2283")
	t.Setenv("ENCRYPTION_KEY", "test-secret")
	for _, tc := range []struct {
		enabled, origin string
		valid           bool
	}{
		{"", "", true},
		{"false", "", true},
		{"true", "", false},
		{"true", "https://places.example", true},
		{"true", "http://localhost:3080", true},
		{"true", "https://places.example/path", false},
		{"true", "https://user:secret@places.example", false},
		{"true", "https://places.example?secret=x", false},
		{"true", "https://places.example#x", false},
		{"true", "https://", false},
		{"true", "null", false},
		{"true", "ftp://places.example", false},
	} {
		t.Run(tc.enabled+"/"+tc.origin, func(t *testing.T) {
			t.Setenv("AI_ENABLED", tc.enabled)
			t.Setenv("AI_PUBLIC_ORIGIN", tc.origin)
			cfg, err := loadConfig()
			if (err == nil) != tc.valid {
				t.Fatalf("valid = %v, error = %v", tc.valid, err)
			}
			if tc.valid && (cfg.AIEnabled != (tc.enabled == "true") || cfg.AIPublicOrigin != tc.origin) {
				t.Fatal("AI configuration was not retained")
			}
		})
	}
}
