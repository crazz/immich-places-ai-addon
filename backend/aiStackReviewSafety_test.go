package main

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestAIStackReviewNeverSubstitutesPartialUnavailableScope(t *testing.T) {
	for _, kind := range []string{"hidden", "trashed", "nonimage", "foreign-stack", "access", "bad-gps", "partial-stack", "oversized-stack", "local-hidden"} {
		t.Run(kind, func(t *testing.T) {
			f := stackWriteFixture(t)
			member := f.metadata[selectionB]
			switch kind {
			case "hidden":
				member["visibility"] = "hidden"
			case "trashed":
				member["isTrashed"] = true
			case "nonimage":
				member["type"] = "VIDEO"
			case "foreign-stack":
				member["stack"] = map[string]any{"id": selectionID(88), "primaryAssetId": selectionB, "assetCount": 1}
			case "bad-gps":
				member["exifInfo"] = map[string]any{"latitude": "not-coordinate"}
			case "partial-stack":
				f.members = f.members[:2]
			case "local-hidden":
				if _, err := f.f.db.db.Exec(`UPDATE assets SET isHidden=1 WHERE userID=? AND immichID=?`, testUserID, selectionB); err != nil {
					t.Fatal(err)
				}
			case "access":
				f.handle = func(w http.ResponseWriter, r *http.Request) bool {
					if r.URL.Path == "/api/assets/"+selectionB {
						w.WriteHeader(403)
						return true
					}
					return false
				}
			case "oversized-stack":
				f.handle = func(w http.ResponseWriter, r *http.Request) bool {
					if strings.HasPrefix(r.URL.Path, "/api/stacks/") {
						_, _ = w.Write([]byte(strings.Repeat(" ", 1<<20+1)))
						return true
					}
					return false
				}
			}
			review, err := f.writer.drafts.observeStack(context.Background(), testUserID, f.draft.ID, f.draft.Revision, []string{f.draft.AssetID, selectionB}, f.image.service)
			if err == nil || review.ID != "" {
				t.Fatal("unavailable target produced partial authority", err)
			}
			var rows int
			if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_stack_reviews`).Scan(&rows); err != nil || rows != 0 {
				t.Fatal("failed review persisted", rows, err)
			}
			if _, bad := f.image.counts(); bad != 0 {
				t.Fatal("review mutated upstream")
			}
		})
	}
}
