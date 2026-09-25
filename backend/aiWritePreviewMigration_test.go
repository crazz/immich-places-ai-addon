package main

import (
	"context"
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestAIWritePreviewUpgradeFrom26RollbackAndPooledConstraints(t *testing.T) {
	f, image, _, draft, _ := writePreviewFixture(t)
	if err := goose.DownTo(f.db.db, "migrations", 26); err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.db, "CREATE TABLE ai_write_previews(sentinel TEXT)")
	if err := runMigrations(f.db.db); err == nil {
		t.Fatal("conflicting migration did not fail")
	}
	version, err := goose.GetDBVersion(f.db.db)
	if err != nil || version != 26 {
		t.Fatal("migration partially applied", version, err)
	}
	selectionSQL(t, f.db, "DROP TABLE ai_write_previews")
	f.reopen(t)
	if err = runMigrations(f.db.db); err != nil {
		t.Fatal(err)
	}
	preview := savedWritePreview(t, f, image, draft)
	ctx := context.Background()
	if _, err = f.store.ReadAnalysis(ctx, testUserID, draft.AnalysisID); err != nil {
		t.Fatal("legacy read incompatible with retained preview tables", err)
	}
	if err = f.db.createUser(ctx, "foreign-preview-owner", "foreign-migration@example.com", "hash"); err != nil {
		t.Fatal(err)
	}
	var conns []*sql.Conn
	defer func() {
		for _, conn := range conns {
			conn.Close()
		}
	}()
	for range 4 {
		conn, err := f.db.db.Conn(ctx)
		if err != nil {
			t.Fatal(err)
		}
		conns = append(conns, conn)
		var enabled int
		if err = conn.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&enabled); err != nil || enabled != 1 {
			t.Fatal("pooled foreign keys disabled", enabled, err)
		}
		_, err = conn.ExecContext(ctx, `INSERT INTO ai_write_previews(userID,installationID,id,draftID,revision,version,payload,digest,createdAt,expiresAt) SELECT 'foreign-preview-owner',installationID,id,draftID,revision,version,payload,digest,createdAt,expiresAt FROM ai_write_previews WHERE id=?`, preview.Plan.ID)
		if err == nil {
			t.Fatal("cross-owner revision reference accepted")
		}
	}
	var violations int
	if err = f.db.db.QueryRow("SELECT count(*) FROM pragma_foreign_key_check").Scan(&violations); err != nil || violations != 0 {
		t.Fatal(violations, err)
	}
}
