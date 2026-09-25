package main

import (
	"context"
	"strings"
	"testing"
)

func TestAIWriteRejectsCorruptStoredOperationBeforeReadOrDispatch(t *testing.T) {
	w := newAIWriteFixture(t)
	if _, err := w.f.db.db.Exec(`DROP TRIGGER ai_write_approval_immutable`); err != nil {
		t.Fatal(err)
	}
	if _, err := w.f.db.db.Exec(`UPDATE ai_write_operations SET digest=?`, strings.Repeat("0", 64)); err != nil {
		t.Fatal(err)
	}
	if _, err := w.writer.get(context.Background(), testUserID, w.op.ID, false); err == nil {
		t.Fatal("corrupt approved plan returned")
	}
	_ = w.writer.runOne(context.Background(), testUserID, w.op.ID)
	if sends, reads := w.counts(); sends != 0 || reads != 0 {
		t.Fatal("corrupt plan used", sends, reads)
	}
}
