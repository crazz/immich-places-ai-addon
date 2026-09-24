package main

import (
	"context"
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestAIDraftUpgradeFrom25AndPoolConstraints(t *testing.T) {
	f, analysis := draftFixture(t)
	if err := goose.DownTo(f.db.db, "migrations", 25); err != nil {
		t.Fatal(err)
	}
	f.reopen(t)
	if err := runMigrations(f.db.db); err != nil {
		t.Fatal(err)
	}
	value := acceptedDraft(t, f, analysis)
	ctx := context.Background()
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
			t.Fatal(enabled, err)
		}
		if _, err = conn.ExecContext(ctx, `INSERT INTO ai_draft_revisions(userID,installationID,draftID,revision,content) VALUES('other',?,?,2,'{}')`, f.store.binding, value.ID); err == nil {
			t.Fatal("foreign revision accepted")
		}
	}
	var count int
	if err := f.db.db.QueryRow("SELECT count(*) FROM pragma_foreign_key_check").Scan(&count); err != nil || count != 0 {
		t.Fatal(count, err)
	}
	if err := f.db.db.QueryRow("SELECT count(*) FROM ai_analyses WHERE id=?", analysis).Scan(&count); err != nil || count != 1 {
		t.Fatal("analysis lost", count, err)
	}
}

func TestAIDraftFailedMigrationRollsBackAndRetainedTablesAllowLegacyReads(t *testing.T) {
	f, analysis := draftFixture(t)
	if err := goose.DownTo(f.db.db, "migrations", 25); err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.db, "CREATE TABLE ai_draft_revisions(sentinel TEXT)")
	if err := runMigrations(f.db.db); err == nil {
		t.Fatal("conflicting migration passed")
	}
	var count int
	if err := f.db.db.QueryRow("SELECT count(*) FROM sqlite_master WHERE name='ai_drafts'").Scan(&count); err != nil || count != 0 {
		t.Fatal("migration partially applied", count, err)
	}
	version, err := goose.GetDBVersion(f.db.db)
	if err != nil || version != 25 {
		t.Fatal(version, err)
	}
	selectionSQL(t, f.db, "DROP TABLE ai_draft_revisions")
	if err = runMigrations(f.db.db); err != nil {
		t.Fatal(err)
	}
	value := acceptedDraft(t, f, analysis)
	legacy, err := f.store.ReadAnalysis(context.Background(), testUserID, analysis)
	if err != nil || legacy.ID != analysis {
		t.Fatal("legacy analysis read failed", err)
	}
	f.reopen(t)
	if got := acceptedDraft(t, f, analysis); got.ID != value.ID || got.Revision != 1 {
		t.Fatal("retained draft changed", got)
	}
}
