package main

import (
	"context"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/providers"
)

func TestCapabilityChecksUpgradeFrom18(t *testing.T) {
	db := openTestSQLite(t)
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}
	if err := goose.UpTo(db, "migrations", 18); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users(ID,email,passwordHash) VALUES ('upgrade','upgrade@example.com','hashed')`); err != nil {
		t.Fatal(err)
	}
	database := &Database{db: db, encryptionKey: deriveKey("upgrade-key")}
	profile, err := database.createAIProvider(context.Background(), "upgrade", "provider", providers.Input{
		Config:  providers.Config{Name: "Upgrade", BaseURL: "https://provider.example/v1", Model: "vision"},
		Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := runMigrations(db); err != nil {
		t.Fatal(err)
	}
	started := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	deadline := started.Add(2 * time.Minute)
	admission, err := database.admitAIProviderCapability(context.Background(), "upgrade", profile.ID, profile.Revision, "policy", started, deadline)
	if err != nil {
		t.Fatal(err)
	}
	completed := started.Add(time.Second)
	report := capabilities.Report{
		AttemptID:         admission.AttemptID,
		ProfileID:         profile.ID,
		Revision:          profile.Revision,
		ProtocolVersion:   capabilities.ProtocolVersion,
		PolicyFingerprint: "policy",
		Lifecycle:         "completed",
		StartedAt:         started,
		DeadlineAt:        deadline,
		CompletedAt:       &completed,
		RequestedModel:    "vision",
		Observations:      capabilities.EmptyObservations(),
		Compatibility:     "incomplete",
	}
	report.Observations.Image.Status = capabilities.StatusSupported
	if err := database.completeAIProviderCapability(context.Background(), "upgrade", report); err != nil {
		t.Fatal(err)
	}
	if err := runMigrations(db); err != nil {
		t.Fatal(err)
	}
	loaded, err := database.loadAIProviderCapability(context.Background(), "upgrade", profile.ID, profile.Revision, capabilities.ApplicabilityContext{})
	if err != nil || loaded == nil || loaded.AttemptID != admission.AttemptID || loaded.Observations.Image.Status != capabilities.StatusSupported {
		t.Fatalf("upgraded capability row missing: %+v err=%v", loaded, err)
	}
	var cipherCount int
	if err := db.QueryRow(`SELECT count(*) FROM ai_provider_versions WHERE secretCiphertext IS NOT NULL`).Scan(&cipherCount); err != nil {
		t.Fatal(err)
	}
	if cipherCount != 0 {
		t.Fatal("upgrade path exposed or invented credentials")
	}
}
