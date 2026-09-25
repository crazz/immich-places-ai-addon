package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestAIWriteLocalFailureRecoversReadOnlyAndMissingRowIsNotFabricated(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(map[bool]string{false: "storage", true: "missing"}[missing], func(t *testing.T) {
			w := newAIWriteFixture(t)
			if missing {
				w.handle = func(out http.ResponseWriter, r *http.Request) bool {
					if r.Method == "PATCH" {
						selectionSQL(t, w.f.db, "DELETE FROM assets WHERE userID=? AND immichID=?", testUserID, w.draft.AssetID)
					}
					return false
				}
			} else {
				selectionSQL(t, w.f.db, `CREATE TRIGGER fail_write_refresh BEFORE UPDATE OF latitude ON assets BEGIN SELECT RAISE(ABORT,'synthetic storage failure'); END`)
			}
			_ = w.writer.runOne(context.Background(), testUserID, w.op.ID)
			op := w.status(t)
			if missing {
				if !op.Verified || op.Refreshed || op.Status != "verifying" || op.Code != "LOCAL_REFRESH_PENDING" {
					t.Fatal(op)
				}
				var n int
				if err := w.f.db.db.QueryRow("SELECT count(*) FROM assets WHERE userID=? AND immichID=?", testUserID, w.draft.AssetID).Scan(&n); err != nil || n != 0 {
					t.Fatal("invented asset", n, err)
				}
				seedAsset(t, w.f.db, w.draft.AssetID, nil, nil, "2026-09-20")
			} else {
				if op.Status != "verifying" {
					t.Fatal(op)
				}
				selectionSQL(t, w.f.db, "DROP TRIGGER fail_write_refresh")
			}
			w.f.now = w.f.now.Add(time.Second)
			w.run(t)
			if op = w.status(t); op.Status != "succeeded" || !op.Refreshed {
				t.Fatal(op)
			}
			if sends, _ := w.counts(); sends != 1 {
				t.Fatal("storage repair resent", sends)
			}
		})
	}
}
