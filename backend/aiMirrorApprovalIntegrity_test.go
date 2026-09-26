package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIMirrorApprovalRejectsExportThatDoesNotMatchSavedRevision(t *testing.T) {
	f := mirrorWriteFixture(t)
	plan := f.preview.Plan
	var exported writepreview.MirrorExport
	if json.Unmarshal(plan.Mirror.Value, &exported) != nil {
		t.Fatal("fixture")
	}
	exported.Descriptions["en"] = "Unreviewed replacement"
	plan.Mirror.Value, _ = json.Marshal(exported)
	plan.ID = uuid.NewString()
	raw, _ := json.Marshal(plan)
	sum := sha256.Sum256(raw)
	digest := hex.EncodeToString(sum[:])
	_, err := f.f.db.db.Exec(`INSERT INTO ai_stack_write_previews(userID,installationID,id,draftID,revision,version,payload,digest,createdAt,expiresAt)
 SELECT userID,installationID,?,draftID,revision,version,?,?,createdAt,expiresAt FROM ai_stack_write_previews WHERE id=?`, plan.ID, string(raw), digest, f.preview.Plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.writer.confirm(context.Background(), testUserID, writeback.Confirmation{PreviewID: plan.ID, Digest: digest, Key: "substitution"}); err == nil {
		t.Fatal("canonical unreviewed export was approved")
	}
	var count int
	if err = f.f.db.db.QueryRow(`SELECT count(*) FROM ai_mirror_write_steps`).Scan(&count); err != nil || count != 0 {
		t.Fatal("rejected approval left state", err)
	}
}
