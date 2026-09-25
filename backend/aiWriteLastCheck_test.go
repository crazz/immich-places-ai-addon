package main

import (
	"context"
	"testing"
	"time"
)

func TestAIWriteRechecksAuthorityAfterReservationBeforeTransport(t *testing.T) {
	for _, change := range []string{"disabled", "expired", "hidden", "key", "installation", "deleted"} {
		t.Run(change, func(t *testing.T) {
			w := newAIWriteFixture(t)
			ctx := context.Background()
			attempt := &aiWriteAttempt{store: w.writer}
			if _, err := attempt.Read(ctx, w.op); err != nil {
				t.Fatal(err)
			}
			if err := attempt.Reserve(ctx, w.op, false); err != nil {
				t.Fatal(err)
			}
			switch change {
			case "disabled":
				w.enabled = false
			case "expired":
				w.f.now = w.f.now.Add(5 * time.Minute)
			case "hidden":
				selectionSQL(t, w.f.db, "UPDATE assets SET isHidden=1 WHERE userID=?", testUserID)
			case "key":
				key := "revoked"
				if err := w.f.db.updateImmichAPIKey(ctx, testUserID, &key); err != nil {
					t.Fatal(err)
				}
			case "installation":
				selectionSQL(t, w.f.db, "UPDATE ai_installation_identity SET id='rotated'")
			case "deleted":
				selectionSQL(t, w.f.db, "DELETE FROM users WHERE ID=?", testUserID)
			}
			outcome := attempt.Send(ctx, w.op)
			if !outcome.Known || outcome.Code != "not_sent" {
				t.Fatal(outcome)
			}
			if sends, _ := w.counts(); sends != 0 {
				t.Fatal("authority fence bypassed", sends)
			}
		})
	}
}
