package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorMigrationExtendsPreviewStorageWithoutChangingActiveV3(t *testing.T) {
	f := stackWriteFixture(t)
	if err := goose.DownTo(f.f.db.db, "migrations", 42); err != nil {
		t.Fatal(err)
	}
	op := approveStackFixture(t, f, f.members)
	ctx := context.Background()
	step, err := writeback.TargetStep(op, op.Plan.TargetID)
	if err != nil {
		t.Fatal(err)
	}
	attempt := &aiStackAttempt{store: f.writer, parent: op, assetID: step.Plan.TargetID}
	if _, err := attempt.Read(ctx, step); err != nil {
		t.Fatal(err)
	}
	if err := attempt.Reserve(ctx, step, false); err != nil {
		t.Fatal(err)
	}
	before, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	old, _ := json.Marshal(before)
	if err := runMigrations(f.f.db.db); err != nil {
		t.Fatal(err)
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	after, err := f.writer.get(ctx, testUserID, op.ID, false)
	current, _ := json.Marshal(after)
	if err != nil || string(old) != string(current) || after.Plan.Mirror != nil {
		t.Fatal("upgrade altered active old approval/audit", err)
	}
	var guards int
	if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards`).Scan(&guards); err != nil || guards != len(f.members) {
		t.Fatal("upgrade changed active exclusions", err)
	}
	rows, err := f.f.db.db.Query(`PRAGMA foreign_key_check`)
	if err != nil {
		t.Fatal(err)
	}
	if rows.Next() {
		t.Fatal("upgrade broke existing foreign keys")
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = f.f.db.db.Exec(`INSERT INTO ai_stack_write_previews(userID,installationID,id,draftID,revision,version,payload,digest,createdAt,expiresAt)
 SELECT userID,installationID,?,draftID,revision,'mirror-preview-v4',payload,digest,createdAt,expiresAt FROM ai_stack_write_previews WHERE id=?`, uuid.NewString(), op.Plan.ID)
	if err != nil {
		t.Fatal("v4 preview storage unavailable after upgrade", err)
	}
}
