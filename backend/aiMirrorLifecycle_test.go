package main

import (
	"context"
	"errors"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorDraftRevisionCancelsOnlySettledUnstartedMirror(t *testing.T) {
	for _, reserved := range []bool{false, true} {
		t.Run(map[bool]string{false: "blocked", true: "reserved"}[reserved], func(t *testing.T) {
			f := mirrorWriteFixture(t)
			mirrorMutationServer(t, f)
			ctx := context.Background()
			op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "lifecycle"})
			if err != nil {
				t.Fatal(err)
			}
			if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
				t.Fatal(err)
			}
			op, err = f.writer.get(ctx, testUserID, op.ID, false)
			if err != nil {
				t.Fatal(err)
			}
			if reserved {
				a := &aiMirrorAttempt{standard: &aiStackAttempt{store: f.writer, parent: op, assetID: f.draft.AssetID}}
				if _, err = a.read(ctx, false); err != nil {
					t.Fatal(err)
				}
				if err = a.reserve(ctx, false); err != nil {
					t.Fatal(err)
				}
			}
			_, err = f.writer.drafts.edit(ctx, testUserID, f.draft.ID, f.draft.Revision, drafts.Edit{State: "rejected"})
			if reserved && !errors.Is(err, drafts.ErrWriteInProgress) {
				t.Fatal("active mirror did not freeze revision", err)
			}
			if !reserved && err != nil {
				t.Fatal(err)
			}
			saved, err := f.writer.get(ctx, testUserID, op.ID, false)
			if err != nil {
				t.Fatal(err)
			}
			var guards int
			if err = f.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards`).Scan(&guards); err != nil {
				t.Fatal(err)
			}
			if reserved && (saved.Mirror.Status != "writing" || guards != 1) {
				t.Fatal("active sender lost guard")
			}
			if !reserved && (saved.Mirror.Status != "canceled" || guards != 0 || !saved.Targets[0].Verified) {
				t.Fatal("edit lost standard result or retained obsolete mirror", saved.Mirror.Status, guards)
			}
		})
	}
}

func TestAIMirrorDeletionRetainsUnknownSenderAndKnownCompletionCleansOnlyItsGuard(t *testing.T) {
	for _, known := range []bool{false, true} {
		t.Run(map[bool]string{false: "unknown", true: "late-known"}[known], func(t *testing.T) {
			f := mirrorWriteFixture(t)
			mirrorMutationServer(t, f)
			ctx := context.Background()
			op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "delete"})
			if err != nil {
				t.Fatal(err)
			}
			if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
				t.Fatal(err)
			}
			op, err = f.writer.get(ctx, testUserID, op.ID, false)
			if err != nil {
				t.Fatal(err)
			}
			a := &aiMirrorAttempt{standard: &aiStackAttempt{store: f.writer, parent: op, assetID: f.draft.AssetID}}
			if _, err = a.read(ctx, false); err != nil {
				t.Fatal(err)
			}
			if err = a.reserve(ctx, false); err != nil {
				t.Fatal(err)
			}
			selectionSQL(t, f.f.db, `DELETE FROM users WHERE ID=?`, testUserID)
			var guards int
			if err = f.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards`).Scan(&guards); err != nil || guards != 1 {
				t.Fatal("deletion discarded possible metadata sender", guards, err)
			}
			if known {
				if err = a.sent(ctx, writeback.Completion{Known: true}); err == nil {
					t.Fatal("deleted approval resurrected")
				}
				if err = f.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards`).Scan(&guards); err != nil || guards != 0 {
					t.Fatal("known sender retained guard", guards, err)
				}
			}
			for _, table := range []string{"ai_mirror_write_steps", "ai_mirror_write_events", "ai_mirror_records"} {
				var count int
				if err = f.f.db.db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&count); err != nil || count != 0 {
					t.Fatal("private mirror data survived", table, err)
				}
			}
		})
	}
}

func TestAIMirrorHistoryNeverReportsStandardOnlySuccess(t *testing.T) {
	f := mirrorWriteFixture(t)
	mirrorMutationServer(t, f)
	ctx := context.Background()
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "history"})
	if err != nil {
		t.Fatal(err)
	}
	for phase := 0; phase < 3; phase++ {
		saved, err := f.writer.get(ctx, testUserID, op.ID, false)
		if err != nil {
			t.Fatal(err)
		}
		page, err := f.writer.history(ctx, testUserID, f.draft.ID, "")
		if err != nil || len(page.Items) != 1 || page.Items[0].Status != saved.Status {
			t.Fatal("history differs from combined operation", phase, saved.Status, page, err)
		}
		if phase < 2 {
			if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
				t.Fatal(err)
			}
		}
	}
}
