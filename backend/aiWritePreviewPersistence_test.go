package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"immich-places-backend/internal/ai/writepreview"
)

func TestAIWritePreviewPersistsCanonicalPlanAndReloadsWithoutRead(t *testing.T) {
	f, image, _, draft, _ := writePreviewFixture(t)
	body, _ := json.Marshal(map[string]any{"draftId": draft.ID, "draftRevision": draft.Revision})
	rec := aiRequest(writePreviewHandler(f, image), "POST", "/ai/write-previews", string(body), aiTestOrigin, true)
	var preview writepreview.Preview
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &preview) != nil {
		t.Fatal(rec.Code, rec.Body.String())
	}
	var raw []byte
	var digest string
	if err := f.db.db.QueryRow(`SELECT payload,digest FROM ai_write_previews WHERE userID=? AND installationID=? AND id=?`, testUserID, f.store.binding, preview.Plan.ID).Scan(&raw, &digest); err != nil {
		t.Fatal("plan was not persisted", err)
	}
	sum := sha256.Sum256(raw)
	if digest != hex.EncodeToString(sum[:]) || digest != preview.Digest || len(raw) > 16<<10 {
		t.Fatal("canonical digest mismatch")
	}
	encoded, _ := json.Marshal(preview.Plan)
	if string(raw) != string(encoded) {
		t.Fatal("stored plan differs from displayed plan")
	}
	if _, err := f.db.db.Exec(`UPDATE ai_write_previews SET payload='{}' WHERE id=?`, preview.Plan.ID); err == nil {
		t.Fatal("plan mutated")
	}
	reads, bad := image.counts()
	f.reopen(t)
	got := aiRequest(writePreviewHandler(f, image), "GET", "/ai/write-previews/"+preview.Plan.ID, "", "", true)
	if got.Code != 200 || got.Body.String() != rec.Body.String() {
		t.Fatal("reloaded plan changed", got.Code, got.Body.String())
	}
	if after, failures := image.counts(); after != reads || failures != bad {
		t.Fatal("reload contacted upstream")
	}
}
