package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"immich-places-backend/internal/ai/drafts"
)

func TestAIStackReviewFencesAuthorityChangesDuringReads(t *testing.T) {
	for _, kind := range []string{"revision", "credential", "installation", "deleted"} {
		t.Run(kind, func(t *testing.T) {
			f := stackWriteFixture(t)
			entered, release := make(chan struct{}), make(chan struct{})
			f.handle = func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/assets/"+selectionB {
					close(entered)
					select {
					case <-release:
					case <-r.Context().Done():
					}
					return false
				}
				return false
			}
			finished := make(chan error, 1)
			go func() {
				_, err := f.writer.drafts.observeStack(context.Background(), testUserID, f.draft.ID, f.draft.Revision, []string{f.draft.AssetID, selectionB}, f.image.service)
				finished <- err
			}()
			select {
			case <-entered:
			case <-time.After(3 * time.Second):
				t.Fatal("metadata did not reach barrier")
			}
			switch kind {
			case "revision":
				if _, err := f.writer.drafts.edit(context.Background(), testUserID, f.draft.ID, f.draft.Revision, drafts.Edit{State: "draft"}); err != nil {
					t.Fatal(err)
				}
			case "credential":
				key := "replacement-key"
				if err := f.f.db.updateImmichAPIKey(context.Background(), testUserID, &key); err != nil {
					t.Fatal(err)
				}
			case "installation":
				if _, err := f.f.db.db.Exec(`UPDATE ai_installation_identity SET id=?`, selectionID(77)); err != nil {
					t.Fatal(err)
				}
			case "deleted":
				if _, err := f.f.db.db.Exec(`DELETE FROM users WHERE ID=?`, testUserID); err != nil {
					t.Fatal(err)
				}
			}
			close(release)
			select {
			case err := <-finished:
				if err == nil {
					t.Fatal("obsolete observation published")
				}
			case <-time.After(3 * time.Second):
				t.Fatal("review did not finish")
			}
			var rows int
			if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_stack_reviews`).Scan(&rows); err != nil || rows != 0 {
				t.Fatal("obsolete private records survived", rows, err)
			}
		})
	}
}
