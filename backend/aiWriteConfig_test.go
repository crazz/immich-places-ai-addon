package main

import (
	"testing"
)

func TestAIWriteDefaultAndRuntimeGatesFailClosed(t *testing.T) {
	withCleanWorkDir(t)
	t.Setenv("IMMICH_URL", "http://127.0.0.1:8090")
	t.Setenv("ENCRYPTION_KEY", "synthetic-encryption-key")
	t.Setenv("AI_WRITE_ENABLED", "")
	t.Setenv("AI_WRITE_PROFILE", "")
	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AIWriteEnabled || cfg.AIWriteProfile != "" {
		t.Fatal("write defaults enabled", cfg.AIWriteEnabled, cfg.AIWriteProfile)
	}
	for _, tc := range []struct {
		global, write bool
		profile       string
		want          bool
	}{{false, true, "immich-v3.2.2", false}, {true, false, "immich-v3.2.2", false}, {true, true, "unsupported", false}, {true, true, "immich-v3.2.2", true}} {
		local := *cfg
		local.DataDir = t.TempDir()
		local.AIEnabled = tc.global
		local.AIWriteEnabled = tc.write
		local.AIWriteProfile = tc.profile
		runtime := newAIWriteRuntime(&aiResultStore{}, nil, nil, &local)
		if runtime.store.available() != tc.want {
			t.Fatal(tc)
		}
		duplicate := newAIWriteRuntime(&aiResultStore{}, nil, nil, &local)
		if duplicate.store.available() {
			t.Fatal("second runtime allowed dispatch")
		}
		if runtime.lock != nil {
			runtime.lock.Close()
		}
		if duplicate.lock != nil {
			duplicate.lock.Close()
		}
	}
}
