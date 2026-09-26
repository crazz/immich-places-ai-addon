package main

import (
	"encoding/json"
	"testing"

	"immich-places-backend/internal/ai/drafts"
)

func TestAIMirrorChoiceUsesExistingPrivateRevisionStorageAcrossReopen(t *testing.T) {
	f, id := draftFixture(t)
	value := acceptedDraft(t, f, id)
	rec := draftPatch(f, value.ID, `"1"`, `{"mirror":{"direction":true,"languages":["en"]}}`)
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &value) != nil || value.Mirror == nil || !value.Mirror.Direction || value.Revision != 2 {
		t.Fatalf("mirror choice not stored: %d %s", rec.Code, rec.Body.String())
	}
	f.reopen(t)
	f.store.enabled = false
	restored := aiRequest(draftHandler(f), "GET", "/ai/drafts/"+value.ID, "", "", true)
	if restored.Code != 200 || restored.Body.String() != rec.Body.String() {
		t.Fatal("local choice lost after disabled-AI reopen")
	}
	var previous drafts.Draft
	var raw []byte
	if err := f.db.db.QueryRow(`SELECT content FROM ai_draft_revisions WHERE userID=? AND installationID=? AND draftID=? AND revision=1`, testUserID, f.store.binding, value.ID).Scan(&raw); err != nil || json.Unmarshal(raw, &previous) != nil || previous.Mirror != nil {
		t.Fatal("old revision gained mirror authority", err)
	}
	stale := draftPatch(f, value.ID, `"1"`, `{"mirror":null}`)
	if stale.Code != 412 {
		t.Fatal("stale selection changed saved revision", stale.Code)
	}
}
