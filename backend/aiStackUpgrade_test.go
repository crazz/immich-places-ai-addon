package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIStackUpgradePreservesV1V2UnknownApprovalsAndExcludesOverlap(t *testing.T) {
	for _, version := range []string{"gps-preview-v1", "standard-preview-v2"} {
		t.Run(version, func(t *testing.T) {
			f := stackWriteFixture(t)
			ctx := context.Background()
			if version == "gps-preview-v1" {
				var err error
				f.draft, err = f.writer.drafts.edit(ctx, testUserID, f.draft.ID, f.draft.Revision, drafts.Edit{Fields: json.RawMessage(`["gps"]`), State: "staged"})
				if err != nil {
					t.Fatal(err)
				}
				f.draft, err = f.writer.drafts.edit(ctx, testUserID, f.draft.ID, f.draft.Revision, drafts.Edit{State: "staged"})
				if err != nil {
					t.Fatal(err)
				}
			}
			session := &aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service}
			preview, err := writepreview.Create(ctx, session, session, session, testUserID, f.draft.ID, f.draft.Revision, uuid.NewString(), f.f.store.now)
			if err != nil || preview.Plan.Version != version {
				t.Fatal("wrong old approval fixture", err)
			}
			old, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "old-single-photo"})
			if err != nil {
				t.Fatal(err)
			}
			a := &aiWriteAttempt{store: f.writer}
			if _, err := a.Read(ctx, old); err != nil {
				t.Fatal(err)
			}
			if err := a.Reserve(ctx, old, false); err != nil {
				t.Fatal(err)
			}
			before, err := f.writer.get(ctx, testUserID, old.ID, false)
			if err != nil {
				t.Fatal(err)
			}
			rawBefore, err := json.Marshal(before)
			if err != nil {
				t.Fatal(err)
			}
			if err := goose.DownTo(f.f.db.db, "migrations", 35); err != nil {
				t.Fatal(err)
			}
			var priorGuards int
			if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards WHERE token=?`, old.ID).Scan(&priorGuards); err != nil || priorGuards != 1 {
				t.Fatal("old unknown guard missing", err)
			}
			if err := runMigrations(f.f.db.db); err != nil {
				t.Fatal(err)
			}
			f.f.reopen(t)
			f.writer.drafts.results.jobs = f.f.store
			after, err := f.writer.get(ctx, testUserID, old.ID, false)
			rawAfter, marshalErr := json.Marshal(after)
			if err != nil || marshalErr != nil || string(rawAfter) != string(rawBefore) || after.Plan.Manifest != nil || after.Attempts != 1 || after.Status != "writing" {
				t.Fatal("upgrade rewrote old bytes, audit or attempts", err, marshalErr)
			}
			review, err := f.writer.drafts.observeStack(ctx, testUserID, f.draft.ID, f.draft.Revision, f.members, f.image.service)
			if err != nil {
				t.Fatal(err)
			}
			stack, err := session.createStack(ctx, testUserID, f.draft.ID, f.draft.Revision, review.ID)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: stack.Plan.ID, Digest: stack.Digest, Key: "new-stack-overlap"}); err == nil || err.Error() != "TARGET_BUSY" {
				t.Fatal("new version overlapped old unknown sender", err)
			}
			var newRows, guards int
			if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_stack_write_operations`).Scan(&newRows); err != nil || newRows != 0 {
				t.Fatal("overlap created new approval", err)
			}
			if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards`).Scan(&guards); err != nil || guards != priorGuards {
				t.Fatal("upgrade or overlap changed exclusions", err)
			}
		})
	}
}
