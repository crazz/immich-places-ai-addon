package main

import "testing"

func TestAIWriteCapabilityConfigurationIsIndependentAndExact(t *testing.T) {
	withCleanWorkDir(t)
	t.Setenv("IMMICH_URL", "http://127.0.0.1:8090")
	t.Setenv("ENCRYPTION_KEY", "synthetic-encryption-key")
	t.Setenv("AI_WRITE_ENABLED", "true")
	t.Setenv("AI_WRITE_PROFILE", "immich-v3.2.2")
	t.Setenv("AI_WRITE_CAPABILITIES", "")
	const installation = "caf12459-abcd-4321-8421-aaccff110033"
	for _, enabled := range []bool{false, true} {
		if enabled {
			t.Setenv("AI_WRITE_CAPABILITIES", `{"version":1,"installation":"`+installation+`","profile":"immich-v3.2.2","evidence":"docs/evidence/synthetic-only.md","capabilities":["description"]}`)
		}
		cfg, err := loadConfig()
		if err != nil {
			t.Fatal(err)
		}
		cfg.DataDir = t.TempDir()
		runtime := newAIWriteRuntime(&aiResultStore{}, nil, nil, cfg)
		if runtime.lock != nil {
			t.Cleanup(func() { runtime.lock.Close() })
		}
		if runtime.store.capabilities.Allows(installation, cfg.AIWriteProfile, "description") != enabled {
			t.Fatal("description activation ignored separate attestation")
		}
	}
	t.Setenv("AI_WRITE_CAPABILITIES", `{"unknown":"private-value"}`)
	if _, err := loadConfig(); err == nil {
		t.Fatal("invalid capability configuration accepted")
	}
}
