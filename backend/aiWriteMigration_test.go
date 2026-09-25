package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/pressly/goose/v3"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func TestAIWriteUpgradeFromPrerequisitesAndPooledOwnerConstraints(t *testing.T) {
	for _, base := range []int64{25, 26, 27} {
		t.Run(fmt.Sprint(base), func(t *testing.T) {
			f, image, store, draft, _ := writePreviewFixture(t)
			ctx := context.Background()
			if err := goose.DownTo(f.db.db, "migrations", base); err != nil {
				t.Fatal(err)
			}
			if base == 27 {
				selectionSQL(t, f.db, "CREATE TABLE ai_write_operations(sentinel TEXT)")
				if err := runMigrations(f.db.db); err == nil {
					t.Fatal("expected atomic migration failure")
				}
				if version, err := goose.GetDBVersion(f.db.db); err != nil || version != 27 {
					t.Fatal(version, err)
				}
				selectionSQL(t, f.db, "DROP TABLE ai_write_operations")
			}
			if err := runMigrations(f.db.db); err != nil {
				t.Fatal(err)
			}
			if err := runMigrations(f.db.db); err != nil {
				t.Fatal(err)
			}
			if version, err := goose.GetDBVersion(f.db.db); err != nil || version != 28 {
				t.Fatal(version, err)
			}
			var err error
			if base == 25 {
				draft, err = store.accept(ctx, testUserID, draft.AnalysisID, nil)
				if err != nil {
					t.Fatal(err)
				}
			}
			observation, err := store.observe(ctx, testUserID, draft.ID, draft.Revision, image.service)
			if err != nil {
				t.Fatal(err)
			}
			draft, err = store.acknowledge(ctx, testUserID, draft.ID, draft.Revision, observation.ID, image.service)
			if err != nil {
				t.Fatal(err)
			}
			draft, err = store.edit(ctx, testUserID, draft.ID, draft.Revision, drafts.Edit{Camera: json.RawMessage(`{"latitude":0,"longitude":12}`), Fields: json.RawMessage(`["gps"]`), State: "staged"})
			if err != nil {
				t.Fatal(err)
			}
			preview := savedWritePreview(t, f, image, draft)
			writer := &aiWriteStore{drafts: store, enabled: func() bool { return true }, profile: "immich-v3.2.2"}
			op, err := writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "upgrade"})
			if err != nil {
				t.Fatal(err)
			}
			f.reopen(t)
			store.results.jobs = f.store
			if restored, err := writer.get(ctx, testUserID, op.ID, false); err != nil || restored.Digest != op.Digest {
				t.Fatal("upgrade approval lost", err)
			}
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
				if _, err = conn.ExecContext(ctx, `INSERT INTO ai_write_targets(userID,installationID,operationID) VALUES('foreign',?,?)`, f.store.binding, op.ID); err == nil {
					t.Fatal("foreign target reference accepted")
				}
			}
		})
	}
}
