package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

func savedWritePreview(t *testing.T, f *aiJobFixture, image *aiImageFixture, draft drafts.Draft) writepreview.Preview {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"draftId": draft.ID, "draftRevision": draft.Revision})
	rec := aiRequest(writePreviewHandler(f, image), "POST", "/ai/write-previews", string(body), aiTestOrigin, true)
	var preview writepreview.Preview
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &preview) != nil {
		t.Fatal(rec.Code, rec.Body.String())
	}
	return preview
}

func TestAIWritePreviewBoundsCapacityAndCleanupProtectsRetainedHistory(t *testing.T) {
	f, image, _, draft, _ := writePreviewFixture(t)
	var first writepreview.Preview
	for range 10 {
		first = savedWritePreview(t, f, image, draft)
	}
	body, _ := json.Marshal(map[string]any{"draftId": draft.ID, "draftRevision": draft.Revision})
	rec := aiRequest(writePreviewHandler(f, image), "POST", "/ai/write-previews", string(body), aiTestOrigin, true)
	if rec.Code != 409 || !strings.Contains(rec.Body.String(), "PREVIEW_CAPACITY") {
		t.Fatal("unbounded active previews", rec.Code, rec.Body.String())
	}
	selectionSQL(t, f.db, "UPDATE ai_write_previews SET protected=1 WHERE id=?", first.Plan.ID)
	for i := 0; i < 120; i++ {
		selectionSQL(t, f.db, `INSERT INTO ai_write_previews(userID,installationID,id,draftID,revision,version,payload,digest,createdAt,expiresAt) SELECT userID,installationID,?,draftID,revision,version,payload,digest,createdAt,expiresAt FROM ai_write_previews WHERE id=?`, fmt.Sprintf("expired-%d", i), first.Plan.ID)
	}
	f.now = f.now.Add(5 * time.Minute)
	_ = savedWritePreview(t, f, image, draft)
	var count int
	if err := f.db.db.QueryRow("SELECT count(*) FROM ai_write_previews").Scan(&count); err != nil || count != 31 {
		t.Fatal("cleanup not bounded to 100", count, err)
	}
	if err := f.db.db.QueryRow("SELECT count(*) FROM ai_write_previews WHERE id=?", first.Plan.ID).Scan(&count); err != nil || count != 1 {
		t.Fatal("protected preview removed", count, err)
	}
	if _, err := f.db.db.Exec("DELETE FROM ai_draft_revisions WHERE userID=? AND installationID=? AND draftID=? AND revision=?", testUserID, f.store.binding, draft.ID, draft.Revision); err == nil {
		t.Fatal("referenced revision pruned")
	}
	if err := f.db.db.QueryRow("SELECT count(*) FROM ai_analyses WHERE id=?", draft.AnalysisID).Scan(&count); err != nil || count != 1 {
		t.Fatal("history removed", count, err)
	}
}

func TestAIWritePreviewExpiresAndInvalidatesWithoutChangingSavedPlan(t *testing.T) {
	for _, transition := range []string{"expire", "edit", "reject", "invalidate"} {
		t.Run(transition, func(t *testing.T) {
			f, image, store, draft, _ := writePreviewFixture(t)
			preview := savedWritePreview(t, f, image, draft)
			expected := "stale"
			switch transition {
			case "expire":
				f.now = f.now.Add(5 * time.Minute)
				expected = "expired"
			case "edit":
				if _, err := store.edit(context.Background(), testUserID, draft.ID, draft.Revision, drafts.Edit{Camera: json.RawMessage(`{"latitude":1,"longitude":2}`)}); err != nil {
					t.Fatal(err)
				}
			case "reject":
				if _, err := store.edit(context.Background(), testUserID, draft.ID, draft.Revision, drafts.Edit{State: "rejected"}); err != nil {
					t.Fatal(err)
				}
			case "invalidate":
				selectionSQL(t, f.db, "UPDATE ai_write_previews SET invalidated=1 WHERE id=?", preview.Plan.ID)
			}
			reads, bad := image.counts()
			rec := aiRequest(writePreviewHandler(f, image), "GET", "/ai/write-previews/"+preview.Plan.ID, "", "", true)
			var restored writepreview.Preview
			if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &restored) != nil || restored.Status != expected {
				t.Fatal("obsolete preview usable", rec.Code, rec.Body.String())
			}
			before, _ := json.Marshal(preview.Plan)
			after, _ := json.Marshal(restored.Plan)
			if string(before) != string(after) || restored.Digest != preview.Digest {
				t.Fatal("read extended or changed plan")
			}
			if nowReads, nowBad := image.counts(); nowReads != reads || nowBad != bad {
				t.Fatal("stored read contacted upstream")
			}
			if transition == "expire" {
				replacement := savedWritePreview(t, f, image, draft)
				if replacement.Plan.ID == preview.Plan.ID || replacement.Plan.ExpiresAt == preview.Plan.ExpiresAt {
					t.Fatal("old plan renewed")
				}
				if nowReads, _ := image.counts(); nowReads != reads+1 {
					t.Fatal("replacement skipped fresh read")
				}
			}
		})
	}
}
