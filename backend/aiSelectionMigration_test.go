package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
	"immich-places-backend/internal/ai/selection"
)

func TestAISelectionMigrationFrom20PreservesCatalogAndConstraints(t *testing.T) {
	dataDir := t.TempDir()
	sqlite, err := sql.Open("sqlite", filepath.Join(dataDir, "immich-places.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer sqlite.Close()
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}
	if err := goose.UpTo(sqlite, "migrations", 20); err != nil {
		t.Fatal(err)
	}
	db := &Database{db: sqlite, encryptionKey: deriveKey("upgrade-key")}
	selectionSQL(t, db, "INSERT INTO users (ID,email,passwordHash) VALUES (?,'upgrade@example.com','hash')", testUserID)
	selectionSQL(t, db, "INSERT INTO assets (userID,immichID,type,originalFileName,fileCreatedAt) VALUES (?,?,'IMAGE','preserved.jpg','2026-09-19')", testUserID, selectionA)
	if err := runMigrations(sqlite); err != nil {
		t.Fatal(err)
	}
	if err := runMigrations(sqlite); err != nil {
		t.Fatal(err)
	}
	if err := sqlite.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = newDatabase(dataDir, "upgrade-key")
	if err != nil {
		t.Fatal(err)
	}
	defer db.close()
	sqlite = db.db
	store := selectionStore(t, db)
	result, err := store.preview(context.Background(), selection.Input{Mode: "explicit", AssetIDs: []string{selectionA}, Scope: &selection.Scope{View: "all"}}, testUserID)
	if err != nil || result.EligibleCount != 1 {
		t.Fatalf("upgraded selection: %+v %v", result, err)
	}
	var filename string
	if err := sqlite.QueryRow("SELECT originalFileName FROM assets WHERE immichID=?", selectionA).Scan(&filename); err != nil || filename != "preserved.jpg" {
		t.Fatalf("catalog changed: %s %v", filename, err)
	}
	if _, err := sqlite.Exec("INSERT INTO ai_selection_items VALUES ('foreign',?,0,'asset','{}')", *result.SnapshotID); err == nil {
		t.Fatal("owner-qualified FK not enforced")
	}
	rows, err := sqlite.Query("PRAGMA foreign_key_check")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if rows.Next() {
		t.Fatal("foreign key violation after upgrade")
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}
