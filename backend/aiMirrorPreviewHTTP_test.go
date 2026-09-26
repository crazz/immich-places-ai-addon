package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIMirrorHTTPPreviewRequiresDisclosureAndKeepsExactStackScope(t *testing.T) {
	for _, stacked := range []bool{false, true} {
		t.Run(fmt.Sprint(stacked), func(t *testing.T) {
			f := stackWriteFixture(t)
			ctx := context.Background()
			f.writer.capabilities.Capabilities = []string{"description", "stack_gps", "metadata"}
			f.handle = func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path != "/api/assets/"+f.draft.AssetID+"/metadata" {
					return false
				}
				if r.Method != "GET" {
					t.Error("preview attempted metadata write")
				}
				_, _ = w.Write([]byte(`[]`))
				return true
			}
			d, err := f.writer.drafts.edit(ctx, testUserID, f.draft.ID, f.draft.Revision, drafts.Edit{Mirror: json.RawMessage(`{"languages":["en"]}`)})
			if err != nil {
				t.Fatal(err)
			}
			d, err = f.writer.drafts.edit(ctx, testUserID, d.ID, d.Revision, drafts.Edit{State: "staged"})
			if err != nil {
				t.Fatal(err)
			}
			reviewID, targets := "", 1
			if stacked {
				review, err := f.writer.drafts.observeStack(ctx, testUserID, d.ID, d.Revision, f.members[:2], f.image.service)
				if err != nil {
					t.Fatal(err)
				}
				reviewID, targets = review.ID, 2
			}
			body := map[string]any{"draftId": d.ID, "draftRevision": d.Revision, "stackReviewId": reviewID}
			raw, _ := json.Marshal(body)
			f.writer.drafts.results.origin = aiTestOrigin
			handler := newAIResultHandler(f.writer.drafts.results, f.image.service)
			without := aiRequest(handler, "POST", "/ai/write-previews", string(raw), aiTestOrigin, true)
			if without.Code == 200 {
				t.Fatal("visibility disclosure omitted")
			}
			body["mirrorDisclosure"] = writepreview.MirrorDisclosure
			raw, _ = json.Marshal(body)
			rec := aiRequest(handler, "POST", "/ai/write-previews", string(raw), aiTestOrigin, true)
			var preview writepreview.Preview
			if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &preview) != nil || preview.Plan.Mirror == nil || preview.Plan.Version != "mirror-preview-v4" {
				t.Fatalf("mirror HTTP preview: %d %s", rec.Code, rec.Body.String())
			}
			if len(preview.Plan.Manifest.Targets) != targets || preview.Plan.TargetID != d.AssetID {
				t.Fatal("wrong exact target manifest")
			}
			for _, target := range preview.Plan.Manifest.Targets {
				if target.AssetID != d.AssetID && (len(target.Fields) != 1 || target.Fields[0] != "gps" || target.Description != nil) {
					t.Fatal("metadata expanded sibling authority")
				}
			}
		})
	}
}
