package main

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIWritePreviewPreservesAbsentPartialAndZeroGPS(t *testing.T) {
	for _, gps := range []any{"absent", nil, map[string]any{}, map[string]any{"latitude": 0}, map[string]any{"longitude": 0}, map[string]any{"latitude": math.Copysign(0, -1), "longitude": 0}} {
		f, image, store, draft, meta := writePreviewFixture(t)
		if gps == "absent" {
			delete(meta, "exifInfo")
		} else {
			meta["exifInfo"] = gps
		}
		ctx := context.Background()
		if gps != "absent" {
			observation, err := store.observe(ctx, testUserID, draft.ID, draft.Revision, image.service)
			if err != nil {
				t.Fatal(err)
			}
			draft, err = store.acknowledge(ctx, testUserID, draft.ID, draft.Revision, observation.ID, image.service)
			if err != nil {
				t.Fatal(err)
			}
			draft, err = store.edit(ctx, testUserID, draft.ID, draft.Revision, drafts.Edit{State: "staged"})
			if err != nil {
				t.Fatal(err)
			}
		}
		body, _ := json.Marshal(map[string]any{"draftId": draft.ID, "draftRevision": draft.Revision})
		rec := aiRequest(writePreviewHandler(f, image), "POST", "/ai/write-previews", string(body), aiTestOrigin, true)
		if rec.Code != 200 {
			t.Fatal("nullable GPS rejected", gps, rec.Code, rec.Body.String())
		}
		var preview writepreview.Preview
		if json.Unmarshal(rec.Body.Bytes(), &preview) != nil || !writepreview.EqualGPS(preview.Plan.Before, writepreview.GPS{Latitude: draft.Baseline.Latitude, Longitude: draft.Baseline.Longitude}) {
			t.Fatal("nullable values changed", rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), ":-0") {
			t.Fatal("signed zero was not canonicalized")
		}
	}
}

func TestAIWritePreviewAlreadyMatchingGPSIsUnchangedAcrossReload(t *testing.T) {
	f, image, store, draft, meta := writePreviewFixture(t)
	meta["exifInfo"] = map[string]any{"latitude": 0, "longitude": 12}
	ctx := context.Background()
	observation, err := store.observe(ctx, testUserID, draft.ID, draft.Revision, image.service)
	if err != nil {
		t.Fatal(err)
	}
	draft, err = store.acknowledge(ctx, testUserID, draft.ID, draft.Revision, observation.ID, image.service)
	if err != nil {
		t.Fatal(err)
	}
	draft, err = store.edit(ctx, testUserID, draft.ID, draft.Revision, drafts.Edit{State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"draftId": draft.ID, "draftRevision": draft.Revision})
	h := writePreviewHandler(f, image)
	rec := aiRequest(h, "POST", "/ai/write-previews", string(body), aiTestOrigin, true)
	var preview writepreview.Preview
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &preview) != nil || preview.Diff != "unchanged" {
		t.Fatal("matching GPS shown as a change", rec.Code, rec.Body.String())
	}
	if got := aiRequest(h, "GET", "/ai/write-previews/"+preview.Plan.ID, "", "", true); got.Body.String() != rec.Body.String() {
		t.Fatal("reload changed diff", got.Body.String())
	}
	if _, bad := image.counts(); bad != 0 {
		t.Fatal("read-only comparison mutated upstream")
	}
}
