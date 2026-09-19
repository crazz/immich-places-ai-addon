package main

import (
	"context"
	"testing"
)

func TestAIProviderStorageMigration(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	for _, table := range []string{"ai_provider_profiles", "ai_provider_versions"} {
		var count int
		if err := db.db.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil {
			t.Fatalf("provider table %s unavailable: %v", table, err)
		}
		if count != 0 {
			t.Errorf("new table %s contains %d rows", table, count)
		}
	}
}

func TestAIProviderConstraintsOnEveryConnection(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		conn, err := db.db.Conn(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.Close() })
		for pragma, want := range map[string]int{"foreign_keys": 1, "busy_timeout": 5000} {
			var got int
			if err := conn.QueryRowContext(ctx, "PRAGMA "+pragma).Scan(&got); err != nil || got != want {
				t.Errorf("connection %d: %s = %d, want %d; error = %v", i, pragma, got, want, err)
			}
		}
		_, err = conn.ExecContext(ctx, "INSERT INTO ai_provider_profiles(userID, id, activeRevision, enabled) VALUES ('nonexistent', 'profile', 1, 1)")
		if err == nil {
			t.Errorf("connection %d accepted an ownerless provider", i)
		}
	}
}
