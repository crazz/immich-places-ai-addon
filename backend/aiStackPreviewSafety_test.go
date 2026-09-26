package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"testing"
)

func TestAIStackPreviewRejectsChangedScopeOrCapabilityDuringRead(t *testing.T) {
	for _, kind := range []string{"gps", "source", "departed", "capability"} {
		t.Run(kind, func(t *testing.T) {
			f := stackWriteFixture(t)
			ctx := context.Background()
			review, err := f.writer.drafts.observeStack(ctx, testUserID, f.draft.ID, f.draft.Revision, []string{f.draft.AssetID, selectionB}, f.image.service)
			if err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "gps":
				f.metadata[selectionB]["exifInfo"] = map[string]any{"latitude": 33, "longitude": 44}
			case "source":
				f.metadata[selectionB]["checksum"] = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{9}, 20))
			case "departed":
				f.members = []string{f.draft.AssetID, selectionID(3)}
				for _, meta := range f.metadata {
					meta["stack"] = map[string]any{"id": f.stackID, "primaryAssetId": f.draft.AssetID, "assetCount": 2}
				}
			case "capability":
				f.handle = func(w http.ResponseWriter, r *http.Request) bool {
					if r.URL.Path == "/api/stacks/"+f.stackID {
						f.writer.capabilities.Evidence = "changed policy"
					}
					return false
				}
			}
			session := &aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service}
			preview, err := session.createStack(ctx, testUserID, f.draft.ID, f.draft.Revision, review.ID)
			if err == nil || preview.Status == "usable" {
				t.Fatal("changed reviewed decision produced a usable preview", err)
			}
			var rows int
			if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_stack_write_previews`).Scan(&rows); err != nil || rows != 0 {
				t.Fatal("changed preview persisted", rows, err)
			}
		})
	}
}
