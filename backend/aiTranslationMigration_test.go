package main

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
	"testing"
	"time"
)

func TestAITranslationUpgradePreservesDraftAndPoolConstraints(t *testing.T) {
	f, s, req := translationFixture(t)
	ctx := context.Background()
	before, err := s.drafts.get(ctx, testUserID, req.DraftID)
	if err != nil {
		t.Fatal(err)
	}
	if err = goose.DownTo(f.db.db, "migrations", 28); err != nil {
		t.Fatal(err)
	}
	f.reopen(t)
	s.drafts.results.jobs = f.store
	if err = runMigrations(f.db.db); err != nil {
		t.Fatal(err)
	}
	run, err := s.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.db.Exec("UPDATE ai_translation_runs SET requestJSON='{}' WHERE id=?", run.ID); err == nil {
		t.Fatal("immutable consent altered")
	}
	var conns []*sql.Conn
	defer func() {
		for _, c := range conns {
			c.Close()
		}
	}()
	for range 4 {
		c, e := f.db.db.Conn(ctx)
		if e != nil {
			t.Fatal(e)
		}
		conns = append(conns, c)
		var enabled int
		if e = c.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&enabled); e != nil || enabled != 1 {
			t.Fatal(enabled, e)
		}
		if _, e = c.ExecContext(ctx, `INSERT INTO ai_translation_items(userID,installationID,runID,language,state) VALUES('foreign',?,?,'en','queued')`, f.store.binding, run.ID); e == nil {
			t.Fatal("foreign item admitted")
		}
	}
	for _, c := range conns {
		c.Close()
	}
	conns = nil
	if _, err = f.store.PurgeBefore(ctx, testUserID, f.store.now().Add(-24*time.Hour), 100); err != nil {
		t.Fatal(err)
	}
	got, err := s.drafts.get(ctx, testUserID, req.DraftID)
	if err != nil || got.Revision != before.Revision {
		t.Fatal("draft lost", err)
	}
	f.reopen(t)
	s.drafts.results.jobs = f.store
	again, err := s.submit(ctx, testUserID, req)
	if err != nil || again.ID != run.ID {
		t.Fatal("history lost", err)
	}
	var violations int
	if err = f.db.db.QueryRow("SELECT count(*) FROM pragma_foreign_key_check").Scan(&violations); err != nil || violations != 0 {
		t.Fatal(violations, err)
	}
}
