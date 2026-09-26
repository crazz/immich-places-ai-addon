package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorRevokedAuthorityBetweenStepsNeverSends(t *testing.T) {
	for _, scenario := range []string{"disabled", "credential", "installation", "source", "namespace", "expired", "capability", "partial-standard"} {
		t.Run(scenario, func(t *testing.T) {
			f := mirrorWriteFixture(t)
			mirrorMutationServer(t, f)
			ctx := context.Background()
			op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "authority"})
			if err != nil {
				t.Fatal(err)
			}
			if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "disabled":
				f.writer.enabled = func() bool { return false }
			case "credential":
				key := "replacement-key"
				if err = f.f.db.updateImmichAPIKey(ctx, testUserID, &key); err != nil {
					t.Fatal(err)
				}
			case "installation":
				selectionSQL(t, f.f.db, `UPDATE ai_installation_identity SET id='replacement' WHERE singleton=1`)
			case "source":
				f.meta["checksum"] = "changed-image"
			case "namespace":
				f.namespace = json.RawMessage(`{"foreign":true}`)
			case "expired":
				f.f.now = f.f.now.Add(6 * time.Minute)
			case "capability":
				f.writer.capabilities.Capabilities = []string{"description"}
			case "partial-standard":
				selectionSQL(t, f.f.db, `UPDATE ai_stack_write_targets SET status='partial' WHERE operationID=?`, op.ID)
			}
			_ = f.writer.runOne(ctx, testUserID, op.ID)
			if f.standardSends != 1 || f.mirrorSends != 0 {
				t.Fatal("revoked metadata authority sent mutation")
			}
			if scenario == "installation" {
				if _, err = f.writer.get(ctx, testUserID, op.ID, false); err == nil {
					t.Fatal("obsolete installation retained private access")
				}
				return
			}
			saved, err := f.writer.get(ctx, testUserID, op.ID, false)
			if err != nil || !saved.Targets[0].Verified {
				t.Fatal("retained standard evidence lost", err)
			}
			if scenario == "partial-standard" && saved.Mirror.Status != "blocked" {
				t.Fatal("incomplete standard did not block mirror")
			}
			if scenario == "expired" && (saved.Mirror.Status != "expired" || saved.Mirror.Code != "APPROVAL_EXPIRED") {
				t.Fatal("expired approval lacks explicit outcome", saved.Mirror.Status, saved.Mirror.Code)
			}
		})
	}
}
