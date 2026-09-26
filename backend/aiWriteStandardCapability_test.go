package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStandardCapabilityChangesCannotAuthorizeConfirmationOrDispatch(t *testing.T) {
	for _, phase := range []string{"confirmation", "dispatch"} {
		for _, change := range []string{"absent", "installation", "profile", "evidence", "disabled"} {
			t.Run(phase+"/"+change, func(t *testing.T) {
				w := standardWriteFixture(t, []string{"description"})
				ctx := context.Background()
				input := writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "capability-fence"}
				var op writeback.Operation
				attempt := &aiWriteAttempt{store: w.writer}
				if phase == "dispatch" {
					var err error
					op, err = w.writer.confirm(ctx, testUserID, input)
					if err != nil {
						t.Fatal(err)
					}
					if _, err = attempt.Read(ctx, op); err != nil {
						t.Fatal(err)
					}
					if err = attempt.Reserve(ctx, op, false); err != nil {
						t.Fatal(err)
					}
				}
				reads, bad := w.image.counts()
				switch change {
				case "absent":
					w.writer.capabilities = writeback.CapabilityPolicy{}
				case "installation":
					w.writer.capabilities.Installation = "wrong-installation"
				case "profile":
					w.writer.capabilities.Profile = "unknown-profile"
				case "evidence":
					w.writer.capabilities.Evidence = "new-unreviewed-attestation"
				case "disabled":
					w.writer.enabled = func() bool { return false }
				}
				if phase == "confirmation" {
					if _, err := w.writer.confirm(ctx, testUserID, input); err == nil {
						t.Fatal("changed capability confirmed old plan")
					}
				} else {
					result := attempt.Send(ctx, op)
					if !result.Known || result.Code != "not_sent" {
						t.Fatal("changed capability reached transport", result)
					}
					if retained, err := w.writer.get(ctx, testUserID, op.ID, false); err != nil || retained.Digest != op.Digest {
						t.Fatal("capability change hid local approval", err)
					}
				}
				if after, invalid := w.image.counts(); after != reads || invalid != bad {
					t.Fatal("denied capability caused upstream I/O")
				}
			})
		}
	}
}
