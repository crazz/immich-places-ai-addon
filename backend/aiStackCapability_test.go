package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStackCapabilityMustRemainExactAtConfirmationAndDispatch(t *testing.T) {
	for _, phase := range []string{"confirmation", "dispatch"} {
		for _, change := range []string{"absent", "description-only", "stack-only", "installation", "profile", "evidence", "disabled"} {
			t.Run(phase+"/"+change, func(t *testing.T) {
				f := stackWriteFixture(t)
				ctx := context.Background()
				review, err := f.writer.drafts.observeStack(ctx, testUserID, f.draft.ID, f.draft.Revision, f.members, f.image.service)
				if err != nil {
					t.Fatal(err)
				}
				preview, err := (&aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service}).createStack(ctx, testUserID, f.draft.ID, f.draft.Revision, review.ID)
				if err != nil {
					t.Fatal(err)
				}
				input := writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "capability-fence"}
				var op, step writeback.Operation
				var attempt *aiStackAttempt
				if phase == "dispatch" {
					op, err = f.writer.confirm(ctx, testUserID, input)
					if err != nil {
						t.Fatal(err)
					}
					step, err = writeback.TargetStep(op, op.Plan.TargetID)
					if err != nil {
						t.Fatal(err)
					}
					attempt = &aiStackAttempt{store: f.writer, parent: op, assetID: op.Plan.TargetID}
					if _, err := attempt.Read(ctx, step); err != nil {
						t.Fatal(err)
					}
					if err := attempt.Reserve(ctx, step, false); err != nil {
						t.Fatal(err)
					}
				}
				reads, bad := f.image.counts()
				switch change {
				case "absent":
					f.writer.capabilities = writeback.CapabilityPolicy{}
				case "description-only":
					f.writer.capabilities.Capabilities = []string{"description"}
				case "stack-only":
					f.writer.capabilities.Capabilities = []string{"stack_gps"}
				case "installation":
					f.writer.capabilities.Installation = "obsolete-installation"
				case "profile":
					f.writer.capabilities.Profile = "unknown-profile"
				case "evidence":
					f.writer.capabilities.Evidence = "new-unreviewed-evidence"
				case "disabled":
					f.writer.enabled = func() bool { return false }
				}
				if phase == "confirmation" {
					if _, err := f.writer.confirm(ctx, testUserID, input); err == nil {
						t.Fatal("changed capability granted stack authority")
					}
					var count int
					if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards`).Scan(&count); err != nil || count != 0 {
						t.Fatal("rejected confirmation reserved targets", err)
					}
				} else {
					if result := attempt.Send(ctx, step); !result.Known || result.Code != "not_sent" {
						t.Fatal("changed capability reached transport", result)
					}
					if saved, err := f.writer.get(ctx, testUserID, op.ID, false); err != nil || saved.Digest != op.Digest {
						t.Fatal("capability loss hid retained audit", err)
					}
				}
				if after, invalid := f.image.counts(); after != reads || invalid != bad {
					t.Fatal("rejected capability performed external I/O")
				}
			})
		}
	}
}
