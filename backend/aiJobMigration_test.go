package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
	"immich-places-backend/internal/ai/selection"
)

func TestAIJobUpgradePreservesDataAndPoolConstraints(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	seedAsset(t, f.db, selectionA, nil, nil, "2026-09-20")
	snapshot, err := f.selection.preview(ctx, selection.Input{Mode: "explicit", AssetIDs: []string{selectionA}, Scope: &selection.Scope{View: "all"}}, testUserID)
	if err != nil || snapshot.SnapshotID == nil {
		t.Fatal("snapshot", err)
	}
	if err = goose.DownTo(f.db.db, "migrations", 21); err != nil {
		t.Fatal(err)
	}
	var sequence int
	var name, path string
	if err = f.db.db.QueryRow("PRAGMA database_list").Scan(&sequence, &name, &path); err != nil {
		t.Fatal(err)
	}
	f.db.close()
	db, err := newDatabase(filepath.Dir(path), "test-encryption-key")
	if err != nil {
		t.Fatal(err)
	}
	defer db.close()
	if err = runMigrations(db.db); err != nil {
		t.Fatal(err)
	}
	for _, query := range []string{"SELECT count(*) FROM assets", "SELECT count(*) FROM ai_provider_profiles", "SELECT count(*) FROM ai_selection_snapshots"} {
		var count int
		if err = db.db.QueryRow(query).Scan(&count); err != nil || count != 1 {
			t.Fatal("old data lost", query, count, err)
		}
	}
	store := newAIJobStore(db, f.selection.binding, true, f.store.now)
	if _, err = store.Submit(ctx, f.input); err != nil {
		t.Fatal("upgraded submission", err)
	}
	var conns []*sql.Conn
	defer func() {
		for _, conn := range conns {
			conn.Close()
		}
	}()
	for range 4 {
		conn, err := db.db.Conn(ctx)
		if err != nil {
			t.Fatal(err)
		}
		conns = append(conns, conn)
		var enabled int
		if err = conn.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&enabled); err != nil || enabled != 1 {
			t.Fatal("pool foreign keys", enabled, err)
		}
		if _, err = conn.ExecContext(ctx, `INSERT INTO ai_job_items(userID,jobID,id,assetID,position) VALUES('foreign','missing','item','asset',0)`); err == nil {
			t.Fatal("foreign item accepted")
		}
	}
	rows, err := db.db.Query("PRAGMA foreign_key_check")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if rows.Next() {
		t.Fatal("foreign key violation")
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
}
