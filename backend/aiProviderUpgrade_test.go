package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestAIProviderUpgradeFrom17(t *testing.T) {
	db := openTestSQLite(t)
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}
	if err := goose.UpTo(db, "migrations", 17); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users(ID,email,passwordHash) VALUES ('upgrade','upgrade@example.com','hashed')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO assets(userID,immichID,type,originalFileName,fileCreatedAt) VALUES ('upgrade','photo','IMAGE','photo.jpg','2024-01-01')`); err != nil {
		t.Fatal(err)
	}
	if err := runMigrations(db); err != nil {
		t.Fatal(err)
	}
	database := &Database{db: db, encryptionKey: deriveKey("upgrade-key")}
	if _, err := database.createAIProvider(context.Background(), "upgrade", "provider", providerInput()); err != nil {
		t.Fatal(err)
	}
	if err := runMigrations(db); err != nil {
		t.Fatal(err)
	}
	var filename string
	if err := db.QueryRow(`SELECT originalFileName FROM assets WHERE userID='upgrade' AND immichID='photo'`).Scan(&filename); err != nil || filename != "photo.jpg" {
		t.Fatalf("catalog not preserved: %q, error = %v", filename, err)
	}
}

func TestAIProviderDatabaseUsesRelativeDataDirectory(t *testing.T) {
	withCleanWorkDir(t)
	db, err := newDatabase("local data?#", "test-key")
	if err != nil {
		t.Fatal(err)
	}
	defer db.close()
	var seq int
	var name, actual string
	if err := db.db.QueryRow("PRAGMA database_list").Scan(&seq, &name, &actual); err != nil {
		t.Fatal(err)
	}
	want, err := filepath.Abs(filepath.Join("local data?#", "immich-places.db"))
	if err != nil || actual != want {
		t.Fatalf("database path = %q, want %q, error = %v", actual, want, err)
	}
}
