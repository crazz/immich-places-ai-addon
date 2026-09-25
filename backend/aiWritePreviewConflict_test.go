package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/drafts"
)

func TestAIWritePreviewGPSConflictRequiresExplicitBaselineReview(t *testing.T) {
	f, image, store, draft, meta := writePreviewFixture(t)
	meta["exifInfo"] = map[string]any{"latitude": 0, "longitude": 7}
	h := writePreviewHandler(f, image)
	body, _ := json.Marshal(map[string]any{"draftId": draft.ID, "draftRevision": draft.Revision})
	rec := aiRequest(h, "POST", "/ai/write-previews", string(body), aiTestOrigin, true)
	if rec.Code != 409 {
		t.Fatal("GPS conflict not detected", rec.Code, rec.Body.String())
	}
	var conflict struct {
		Code     string
		Conflict struct {
			Before   struct{ Latitude, Longitude *float64 }
			Current  drafts.Point
			Proposed drafts.Point
		}
	}
	if json.Unmarshal(rec.Body.Bytes(), &conflict) != nil || conflict.Code != "IMMICH_CONFLICT" || conflict.Conflict.Before.Latitude != nil || conflict.Conflict.Before.Longitude != nil || conflict.Conflict.Current != (drafts.Point{Latitude: 0, Longitude: 7}) || conflict.Conflict.Proposed != *draft.Camera {
		t.Fatal("missing owned comparison", rec.Body.String())
	}
	var count int
	if err := f.db.db.QueryRow(`SELECT count(*) FROM ai_write_previews`).Scan(&count); err != nil || count != 0 {
		t.Fatal("conflict published plan", count, err)
	}
	ctx := context.Background()
	saved, err := store.get(ctx, testUserID, draft.ID)
	if err != nil || saved.Revision != draft.Revision || saved.Baseline.Longitude != nil {
		t.Fatal("silently accepted baseline")
	}
	observation, err := store.observe(ctx, testUserID, draft.ID, draft.Revision, image.service)
	if err != nil {
		t.Fatal(err)
	}
	renewed, err := store.acknowledge(ctx, testUserID, draft.ID, draft.Revision, observation.ID, image.service)
	if err != nil || renewed.State != "draft" {
		t.Fatal("review did not unstage", err)
	}
	renewed, err = store.edit(ctx, testUserID, draft.ID, renewed.Revision, drafts.Edit{State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	body, _ = json.Marshal(map[string]any{"draftId": draft.ID, "draftRevision": renewed.Revision})
	if got := aiRequest(h, "POST", "/ai/write-previews", string(body), aiTestOrigin, true); got.Code != 200 {
		t.Fatal("explicit recovery failed", got.Code, got.Body.String())
	}
}

func TestAIWritePreviewDistinguishesMaterialImageFromMetadataChanges(t *testing.T) {
	f, image, store, draft, meta := writePreviewFixture(t)
	body, _ := json.Marshal(map[string]any{"draftId": draft.ID, "draftRevision": draft.Revision})
	h := writePreviewHandler(f, image)
	meta["updatedAt"] = "2026-09-25T08:00:00Z"
	meta["description"] = "Unrelated external correction"
	if rec := aiRequest(h, "POST", "/ai/write-previews", string(body), aiTestOrigin, true); rec.Code != 200 {
		t.Fatal("metadata-only change blocked GPS", rec.Code, rec.Body.String())
	}
	meta["checksum"] = base64.StdEncoding.EncodeToString([]byte(strings.Repeat("x", 20)))
	rec := aiRequest(h, "POST", "/ai/write-previews", string(body), aiTestOrigin, true)
	if rec.Code != 409 || !strings.Contains(rec.Body.String(), "SOURCE_CHANGED") {
		t.Fatal("replaced image accepted", rec.Code, rec.Body.String())
	}
	saved, err := store.get(context.Background(), testUserID, draft.ID)
	if err != nil || saved.Baseline.ImageIdentity != draft.Baseline.ImageIdentity || saved.OriginalSourceDigest != draft.OriginalSourceDigest {
		t.Fatal("review or original provenance replaced", err)
	}
}
