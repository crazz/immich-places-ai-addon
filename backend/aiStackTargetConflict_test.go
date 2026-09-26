package main

import (
	"context"
	"testing"
)

func TestAIStackChangedMemberStopsIndependentlyWithoutExpandingManifest(t *testing.T) {
	for _, change := range []string{"gps", "source", "departure", "hidden"} {
		t.Run(change, func(t *testing.T) {
			f := stackWriteFixture(t)
			op := approveStackFixture(t, f, f.members)
			m := stackMutationServer(t, f)
			ctx := context.Background()
			if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
				t.Fatal(err)
			}
			badID := op.Plan.Manifest.Targets[1].AssetID
			switch change {
			case "gps":
				f.metadata[badID]["exifInfo"].(map[string]any)["latitude"] = 63
			case "source":
				f.metadata[badID]["checksum"] = "AgICAgICAgICAgICAgICAgICAgI="
			case "departure":
				f.metadata[badID]["stack"] = nil
			case "hidden":
				f.metadata[badID]["visibility"] = "hidden"
			}
			addStackFixtureMember(t, f, selectionID(4))
			if change == "departure" {
				f.metadata[badID]["stack"] = nil
			}
			for range 3 {
				if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
					t.Fatal(err)
				}
			}
			saved, err := f.writer.get(ctx, testUserID, op.ID, false)
			if err != nil || saved.Status != "partial" || len(saved.Targets) != 3 || saved.Digest != op.Digest {
				t.Fatal("manifest or honest partial outcome changed", err)
			}
			for _, target := range saved.Targets {
				if target.AssetID == badID {
					if target.Attempts != 0 || m.sent(badID) != 0 || (target.Status != "conflict" && target.Status != "failed") {
						t.Fatal("changed member was overwritten")
					}
					if change == "departure" && (target.Status != "conflict" || target.Code != "STACK_MEMBERSHIP_CHANGED") {
						t.Fatal("departed member did not conflict", target.Status, target.Code)
					}
				} else if target.Status != "succeeded" || m.sent(target.AssetID) != 1 {
					t.Fatal("valid sibling could not complete independently")
				}
			}
			if m.sent(selectionID(4)) != 0 {
				t.Fatal("new member entered approval")
			}
		})
	}
}
