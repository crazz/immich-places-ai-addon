package main

import (
	"encoding/json"
	"testing"

	"immich-places-backend/internal/ai/writepreview"
)

func TestAIStackHTTPRequiresExplicitReviewedSelection(t *testing.T) {
	f := stackWriteFixture(t)
	f.writer.drafts.results.origin = aiTestOrigin
	h := newAIResultHandler(f.writer.drafts.results, f.image.service)
	body, err := json.Marshal(map[string]any{"draftId": f.draft.ID, "draftRevision": f.draft.Revision, "targetIds": []string{f.draft.AssetID, selectionB}})
	if err != nil {
		t.Fatal(err)
	}
	rec := aiRequest(h, "POST", "/ai/stack-reviews", string(body), aiTestOrigin, true)
	var review writepreview.StackReview
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &review) != nil || len(review.Targets) != 2 {
		t.Fatal("explicit stack review unavailable", rec.Code, rec.Body.String())
	}
	body, err = json.Marshal(map[string]any{"draftId": f.draft.ID, "draftRevision": f.draft.Revision, "stackReviewId": review.ID})
	if err != nil {
		t.Fatal(err)
	}
	rec = aiRequest(h, "POST", "/ai/write-previews", string(body), aiTestOrigin, true)
	var preview writepreview.Preview
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &preview) != nil || preview.Plan.Manifest == nil || len(preview.Plan.Manifest.Targets) != 2 || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("reviewed preview unavailable", rec.Code, rec.Body.String())
	}
	if _, bad := f.image.counts(); bad != 0 {
		t.Fatal("HTTP review caused a mutation")
	}
}

func TestAIStackHTTPRejectsInvalidSelectionBeforeExternalReads(t *testing.T) {
	f := stackWriteFixture(t)
	f.writer.drafts.results.origin = aiTestOrigin
	h := newAIResultHandler(f.writer.drafts.results, f.image.service)
	many := make([]string, 51)
	for i := range many {
		many[i] = selectionID(i + 1)
	}
	for _, tc := range []struct {
		ids    []string
		origin string
		auth   bool
		want   int
	}{
		{[]string{"malformed"}, aiTestOrigin, true, 400},
		{many, aiTestOrigin, true, 400},
		{nil, "https://foreign.invalid", true, 403},
		{nil, aiTestOrigin, false, 401},
	} {
		body, err := json.Marshal(map[string]any{"draftId": f.draft.ID, "draftRevision": f.draft.Revision, "targetIds": tc.ids})
		if err != nil {
			t.Fatal(err)
		}
		before, _ := f.image.counts()
		rec := aiRequest(h, "POST", "/ai/stack-reviews", string(body), tc.origin, tc.auth)
		after, _ := f.image.counts()
		if rec.Code != tc.want || before != after {
			t.Fatalf("invalid selection crossed read boundary: %d want %d, reads %d→%d", rec.Code, tc.want, before, after)
		}
	}
	body, err := json.Marshal(map[string]any{"draftId": f.draft.ID, "draftRevision": f.draft.Revision, "stackReviewId": selectionID(99)})
	if err != nil {
		t.Fatal(err)
	}
	before, _ := f.image.counts()
	rec := aiRequest(h, "POST", "/ai/write-previews", string(body), aiTestOrigin, true)
	after, _ := f.image.counts()
	if rec.Code < 400 || before != after {
		t.Fatal("unreviewed reference crossed upstream boundary", rec.Code)
	}
}
