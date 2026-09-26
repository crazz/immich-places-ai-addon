package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/writeback"
)

func TestAIStandardConfirmationRejectsUnknownAlteredAndOversizedPlans(t *testing.T) {
	for _, change := range []string{"version", "field", "text", "policy", "owner", "installation", "digest", "unknown-key", "oversize"} {
		t.Run(change, func(t *testing.T) {
			w := standardWriteFixture(t, []string{"description"})
			plan := w.preview.Plan
			plan.ID = uuid.NewString()
			switch change {
			case "version":
				plan.Version = "standard-preview-v3"
			case "field":
				plan.Fields = []string{"description", "heading"}
			case "text":
				plan.Description.Intended = "Unapproved text"
			case "policy":
				plan.PolicyID = strings.Repeat("b", 64)
			case "owner":
				plan.Owner = "foreign-owner"
			case "installation":
				plan.Installation = uuid.NewString()
			}
			raw, err := json.Marshal(plan)
			if err != nil {
				t.Fatal(err)
			}
			if change == "unknown-key" {
				raw = append([]byte(`{"unauthorized":true,`), raw[1:]...)
			}
			if change == "oversize" {
				raw = append(raw, []byte(strings.Repeat(" ", 1<<20))...)
			}
			sum := sha256.Sum256(raw)
			digest := hex.EncodeToString(sum[:])
			if change == "digest" {
				digest = strings.Repeat("0", 64)
			}
			_, err = w.f.db.db.Exec(`INSERT INTO ai_standard_write_previews(userID,installationID,id,draftID,revision,version,payload,digest,createdAt,expiresAt) SELECT userID,installationID,?,draftID,revision,version,?,?,createdAt,expiresAt FROM ai_standard_write_previews WHERE id=?`, plan.ID, string(raw), digest, w.preview.Plan.ID)
			if change == "oversize" {
				if err == nil {
					t.Fatal("oversized durable plan admitted")
				}
				if _, err := writeback.DecodePlan(raw, digest, testUserID, w.f.store.binding, plan.ID); err == nil {
					t.Fatal("oversized plan decoded")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			reads, bad := w.image.counts()
			if _, err := w.writer.confirm(context.Background(), testUserID, writeback.Confirmation{PreviewID: plan.ID, Digest: digest, Key: "invalid-authority"}); err == nil {
				t.Fatal("invalid durable plan confirmed")
			}
			var operations, guards int
			if err := w.f.db.db.QueryRow(`SELECT count(*) FROM ai_standard_write_operations`).Scan(&operations); err != nil {
				t.Fatal(err)
			}
			if err := w.f.db.db.QueryRow(`SELECT count(*) FROM ai_write_target_guards`).Scan(&guards); err != nil || operations != 0 || guards != 0 {
				t.Fatal("rejected authority left operation or guard", err)
			}
			if after, invalid := w.image.counts(); after != reads || invalid != bad {
				t.Fatal("rejected plan caused upstream I/O")
			}
		})
	}
}
