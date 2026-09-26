package main

import (
	"context"
	"testing"

	"github.com/pressly/goose/v3"
	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorLifecycleDowngradeCannotDiscardRetainedStepProtection(t *testing.T) {
	f := mirrorWriteFixture(t)
	ctx := context.Background()
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "downgrade-guard"})
	if err != nil {
		t.Fatal(err)
	}
	if err = goose.DownTo(f.f.db.db, "migrations", 44); err == nil {
		t.Fatal("retained metadata lost lifecycle protections")
	}
	if version, err := goose.GetDBVersion(f.f.db.db); err != nil || version != 45 {
		t.Fatal("failed rollback changed schema version", version, err)
	}
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || saved.Mirror == nil || saved.Mirror.Status != "blocked" {
		t.Fatal("failed rollback damaged retained approval", err)
	}
}
