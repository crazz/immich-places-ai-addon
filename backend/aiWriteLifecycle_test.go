package main

import (
	"testing"
)

func TestAIWriteDisableCancelsQueuedAuthorityWithoutReenableSend(t *testing.T) {
	w := newAIWriteFixture(t)
	w.enabled = false
	w.run(t)
	if op := w.status(t); op.Status != "canceled" {
		t.Fatal(op)
	}
	w.enabled = true
	w.run(t)
	if sends, _ := w.counts(); sends != 0 {
		t.Fatal("reenabled stale approval", sends)
	}
}

func TestAIWriteAccountDeletionKeepsOnlyUncertainOpaqueGuard(t *testing.T) {
	w := newAIWriteFixture(t)
	selectionSQL(t, w.f.db, "DELETE FROM users WHERE ID=?", testUserID)
	var guards int
	if err := w.f.db.db.QueryRow("SELECT count(*) FROM ai_write_target_guards").Scan(&guards); err != nil || guards != 0 {
		t.Fatal("undispatched guard survived deletion", guards, err)
	}
	for _, table := range []string{"ai_write_operations", "ai_write_targets", "ai_write_events", "ai_write_previews"} {
		var n int
		if err := w.f.db.db.QueryRow("SELECT count(*) FROM " + table).Scan(&n); err != nil || n != 0 {
			t.Fatal("private data survived", table, n, err)
		}
	}
}
