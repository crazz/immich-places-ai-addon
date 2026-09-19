package main

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/providers"
)

func TestUsageMigrationPreservesOldEvidenceAndSurvivesReopen(t *testing.T) {
	sqlite := openTestSQLite(t)
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}
	if err := goose.UpTo(sqlite, "migrations", 19); err != nil {
		t.Fatal(err)
	}
	if _, err := sqlite.Exec(`INSERT INTO users(ID,email,passwordHash) VALUES ('upgrade','upgrade@example.com','hashed')`); err != nil {
		t.Fatal(err)
	}
	db := &Database{db: sqlite, encryptionKey: deriveKey("upgrade-key")}
	profile, err := db.createAIProvider(context.Background(), "upgrade", "provider", providers.Input{
		Config: providers.Config{Name: "Upgrade", BaseURL: "https://provider.example/v1", Model: "vision"}, Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = sqlite.Exec(`INSERT INTO ai_provider_capability_checks
		(userID,profileID,revision,attemptID,protocolVersion,policyFingerprint,lifecycle,startedAt,deadlineAt,requestedModel,observationsJSON,compatibility)
		VALUES ('upgrade',?,1,'old-attempt','capability-v1','policy','completed',?,?, 'vision',?, 'incomplete')`,
		profile.ID, time.Now().UTC().Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano),
		`{"image":{"status":"supported"},"json":{"status":"unverified"},"strict":{"status":"unverified"}}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := runMigrations(sqlite); err != nil {
		t.Fatal(err)
	}
	old, err := db.loadAIProviderCapability(context.Background(), "upgrade", profile.ID, 1, capabilities.ApplicabilityContext{AIEnabled: true, ProfileEnabled: true, CurrentPolicyFingerprint: "policy"})
	if err != nil || old == nil || old.AttemptID != "old-attempt" || old.Applicable || old.Usage != nil || !old.InputMayBeConsumed {
		t.Fatalf("migration changed historical evidence or invented usage: %+v %v", old, err)
	}
	old.Usage = &capabilities.Usage{}
	value := 36
	old.Usage.TotalTokens = &value
	if err := db.completeAIProviderCapability(context.Background(), "upgrade", *old); err != nil {
		t.Fatal(err)
	}
	// Open an independent SQLite connection to the same file, without cached report state.
	var seq int
	var name, path string
	if err := sqlite.QueryRow(`PRAGMA database_list`).Scan(&seq, &name, &path); err != nil {
		t.Fatal(err)
	}
	if err := sqlite.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { reopened.Close() })
	db.db = reopened
	stored, err := db.loadAIProviderCapability(context.Background(), "upgrade", profile.ID, 1, capabilities.ApplicabilityContext{})
	if err != nil || stored == nil || stored.Usage == nil || stored.Usage.TotalTokens == nil || *stored.Usage.TotalTokens != 36 || !stored.InputMayBeConsumed {
		t.Fatalf("usage did not survive reopen: %+v %v", stored, err)
	}
}
