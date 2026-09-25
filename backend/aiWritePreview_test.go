package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/drafts"
)

func writePreviewFixture(t *testing.T) (*aiJobFixture, *aiImageFixture, *aiDraftStore, drafts.Draft, map[string]any) {
	t.Helper()
	f, image, store, value, meta := draftBaselineFixture(t)
	ctx := context.Background()
	observation, err := store.observe(ctx, testUserID, value.ID, value.Revision, image.service)
	if err != nil {
		t.Fatal(err)
	}
	value, err = store.acknowledge(ctx, testUserID, value.ID, value.Revision, observation.ID, image.service)
	if err != nil {
		t.Fatal(err)
	}
	value, err = store.edit(ctx, testUserID, value.ID, value.Revision, drafts.Edit{Camera: json.RawMessage(`{"latitude":0,"longitude":12}`), Fields: json.RawMessage(`["gps"]`), State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	return f, image, store, value, meta
}

func TestAIWritePreviewRejectsIncompleteExpandedAndObsoleteInput(t *testing.T) {
	f, image, store, draft, _ := writePreviewFixture(t)
	ctx := context.Background()
	value, err := store.edit(ctx, testUserID, draft.ID, draft.Revision, drafts.Edit{State: "draft"})
	if err != nil {
		t.Fatal(err)
	}
	h := writePreviewHandler(f, image)
	body := func(id string, revision int) string {
		raw, _ := json.Marshal(map[string]any{"draftId": id, "draftRevision": revision})
		return string(raw)
	}
	reads, bad := image.counts()
	for _, tc := range []struct {
		body   string
		status int
		code   string
	}{
		{body(value.ID, value.Revision), 409, "DRAFT_NOT_STAGED"},
		{body(value.ID, draft.Revision), 409, "DRAFT_CONFLICT"},
		{strings.TrimSuffix(body(value.ID, value.Revision), "}") + `,"targetId":"other"}`, 400, "INVALID_PREVIEW"},
		{`{"draftId":"` + value.ID + `","draftRevision":0}`, 400, "INVALID_PREVIEW"},
		{`{"draftId":"` + value.ID + `","draftRevision":4,"draftRevision":4}`, 400, "INVALID_PREVIEW"},
		{strings.Repeat(" ", 4097) + body(value.ID, value.Revision), 400, "INVALID_PREVIEW"},
		{body(selectionB, 1), 404, "PREVIEW_UNAVAILABLE"},
	} {
		rec := aiRequest(h, "POST", "/ai/write-previews", tc.body, aiTestOrigin, true)
		if rec.Code != tc.status || !strings.Contains(rec.Body.String(), tc.code) {
			t.Fatalf("got %d %s; want %d %s", rec.Code, rec.Body.String(), tc.status, tc.code)
		}
	}
	if after, failures := image.counts(); after != reads || failures != bad {
		t.Fatal("invalid input read source")
	}
	value, err = store.edit(ctx, testUserID, value.ID, value.Revision, drafts.Edit{State: "rejected"})
	if err != nil {
		t.Fatal(err)
	}
	if rec := aiRequest(h, "POST", "/ai/write-previews", body(value.ID, value.Revision), aiTestOrigin, true); rec.Code != 409 {
		t.Fatal("rejected draft previewed", rec.Code)
	}
	retained, err := store.get(ctx, testUserID, value.ID)
	if err != nil || retained.Revision != value.Revision {
		t.Fatal("preview changed draft", err)
	}
}

func writePreviewHandler(f *aiJobFixture, image *aiImageFixture) http.Handler {
	return newAIResultHandler(&aiResultStore{jobs: f.store, origin: aiTestOrigin}, image.service)
}

func TestAIWritePreviewExactGPSWithoutQualityGateOrMutation(t *testing.T) {
	f, image, _, draft, _ := writePreviewFixture(t)
	reads, bad := image.counts()
	body, _ := json.Marshal(map[string]any{"draftId": draft.ID, "draftRevision": draft.Revision})
	rec := aiRequest(writePreviewHandler(f, image), "POST", "/ai/write-previews", string(body), aiTestOrigin, true)
	if rec.Code != 200 {
		t.Fatalf("preview: %d %s", rec.Code, rec.Body.String())
	}
	var reply struct {
		Status string `json:"status"`
		Diff   string `json:"diff"`
		Digest string `json:"digest"`
		Plan   struct {
			TargetID      string                                 `json:"targetId"`
			DraftRevision int                                    `json:"draftRevision"`
			Fields        []string                               `json:"fields"`
			Before        struct{ Latitude, Longitude *float64 } `json:"before"`
			Intended      drafts.Point                           `json:"intended"`
		} `json:"plan"`
	}
	if json.Unmarshal(rec.Body.Bytes(), &reply) != nil {
		t.Fatal("invalid response")
	}
	if reply.Status != "usable" || reply.Diff != "changed" || len(reply.Digest) != 64 || reply.Plan.TargetID != draft.AssetID || reply.Plan.DraftRevision != draft.Revision || len(reply.Plan.Fields) != 1 || reply.Plan.Fields[0] != "gps" || reply.Plan.Before.Latitude != nil || reply.Plan.Before.Longitude != nil || reply.Plan.Intended != *draft.Camera {
		t.Fatalf("wrong exact comparison: %+v", reply)
	}
	if afterReads, afterBad := image.counts(); afterReads != reads+1 || afterBad != bad {
		t.Fatalf("unexpected upstream calls %d/%d -> %d/%d", reads, bad, afterReads, afterBad)
	}
}
