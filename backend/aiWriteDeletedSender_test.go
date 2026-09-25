package main

import (
	"context"
	"net/http"
	"testing"
)

func TestAIWriteDeletedOwnerNeverReappearsAfterSenderCompletion(t *testing.T) {
	for _, known := range []bool{true, false} {
		t.Run(map[bool]string{true: "completed", false: "unknown"}[known], func(t *testing.T) {
			w := newAIWriteFixture(t)
			if err := w.f.db.createUser(context.Background(), "other-owner", "other@example.com", "hash"); err != nil {
				t.Fatal(err)
			}
			entered, release := make(chan struct{}), make(chan struct{})
			w.handle = func(out http.ResponseWriter, r *http.Request) bool {
				if r.Method != "PATCH" {
					return false
				}
				close(entered)
				<-release
				if !known {
					conn, _, _ := out.(http.Hijacker).Hijack()
					conn.Close()
					return true
				}
				return false
			}
			done := make(chan error, 1)
			go func() { done <- w.writer.runOne(context.Background(), testUserID, w.op.ID) }()
			<-entered
			selectionSQL(t, w.f.db, "DELETE FROM users WHERE ID=?", testUserID)
			close(release)
			<-done
			for _, table := range []string{"ai_write_operations", "ai_write_targets", "ai_write_events", "ai_write_previews", "ai_drafts"} {
				var n int
				if err := w.f.db.db.QueryRow("SELECT count(*) FROM "+table+" WHERE userID=?", testUserID).Scan(&n); err != nil || n != 0 {
					t.Fatal(table, n, err)
				}
			}
			var guards, other int
			if err := w.f.db.db.QueryRow("SELECT count(*) FROM ai_write_target_guards").Scan(&guards); err != nil {
				t.Fatal(err)
			}
			want := 1
			if known {
				want = 0
			}
			if guards != want {
				t.Fatal("incorrect opaque guard lifecycle", guards, want)
			}
			if err := w.f.db.db.QueryRow("SELECT count(*) FROM users WHERE ID='other-owner'").Scan(&other); err != nil || other != 1 {
				t.Fatal("other owner lost", other, err)
			}
			if sends, _ := w.counts(); sends != 1 {
				t.Fatal(sends)
			}
		})
	}
}
