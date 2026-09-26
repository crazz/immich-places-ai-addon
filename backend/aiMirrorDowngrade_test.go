package main

import (
	"testing"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
)

func TestAIMirrorDowngradeRefusesToDiscardRetainedState(t *testing.T) {
	for _, kind := range []string{"preview", "record"} {
		t.Run(kind, func(t *testing.T) {
			f := stackWriteFixture(t)
			op := approveStackFixture(t, f, f.members)
			var table string
			if kind == "preview" {
				table = "ai_stack_write_previews"
				_, err := f.f.db.db.Exec(`INSERT INTO ai_stack_write_previews(userID,installationID,id,draftID,revision,version,payload,digest,createdAt,expiresAt)
 SELECT userID,installationID,?,draftID,revision,'mirror-preview-v4',payload,digest,createdAt,expiresAt FROM ai_stack_write_previews WHERE id=?`, uuid.NewString(), op.Plan.ID)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				table = "ai_mirror_records"
				_, err := f.f.db.db.Exec(`INSERT INTO ai_mirror_records(userID,installationID,assetID,recordID) VALUES(?,?,?,?)`, testUserID, f.f.store.binding, f.draft.AssetID, uuid.NewString())
				if err != nil {
					t.Fatal(err)
				}
			}
			var before, after int
			if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&before); err != nil {
				t.Fatal(err)
			}
			if err := goose.DownTo(f.f.db.db, "migrations", 42); err == nil {
				t.Fatal("downgrade discarded retained metadata state")
			}
			if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&after); err != nil || before != after {
				t.Fatal("failed downgrade was not atomic", err)
			}
			if err := runMigrations(f.f.db.db); err != nil {
				t.Fatal(err)
			}
		})
	}
}
