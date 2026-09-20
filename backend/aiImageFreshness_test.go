package main

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"immich-places-backend/internal/ai/images"
)

func TestAIImageRejectsLocalAuthorityChangesBeforePublication(t *testing.T) {
	for _, stage := range []string{"before-preview", "during-preview"} {
		for _, change := range []string{"account", "credential", "hidden", "installation", "eligible-library-change"} {
			t.Run(stage+"/"+change, func(t *testing.T) {
				f := newAIImageFixture(t)
				entered, release := make(chan struct{}), make(chan struct{})
				var closeRelease, block sync.Once
				unblock := func() { closeRelease.Do(func() { close(release) }) }
				defer unblock()
				f.handle = func(w http.ResponseWriter, r *http.Request) bool {
					target := "/api/assets/" + selectionA
					if stage == "during-preview" {
						target += "/thumbnail"
					}
					if r.URL.Path == target {
						block.Do(func() {
							close(entered)
							select {
							case <-release:
							case <-r.Context().Done():
							}
						})
					}
					return false
				}
				type result struct {
					image *images.Prepared
					err   error
				}
				done := make(chan result, 1)
				go func() {
					p, err := f.service.prepare(context.Background(), testUserID, f.store.binding, selectionA)
					done <- result{p, err}
				}()
				select {
				case <-entered:
				case <-time.After(3 * time.Second):
					t.Fatal("retrieval barrier not reached")
				}
				switch change {
				case "account":
					selectionSQL(t, f.db, "DELETE FROM users WHERE ID=?", testUserID)
				case "credential":
					key := "new-key"
					if err := f.db.updateImmichAPIKey(context.Background(), testUserID, &key); err != nil {
						t.Fatal(err)
					}
				case "hidden":
					selectionSQL(t, f.db, "UPDATE assets SET isHidden=1")
				case "installation":
					selectionSQL(t, f.db, "UPDATE ai_installation_identity SET id=?", selectionID(90))
				case "eligible-library-change":
					selectionSQL(t, f.db, "UPDATE assets SET libraryID='another-visible-library'")
				}
				unblock()
				var got result
				select {
				case got = <-done:
				case <-time.After(3 * time.Second):
					t.Fatal("preparation did not finish")
				}
				if got.err == nil || got.image != nil {
					if got.image != nil {
						got.image.Release()
					}
					t.Fatal("stale authority published a copy")
				}
				calls, _ := f.counts()
				want := 1
				if stage == "during-preview" {
					want = 2
				}
				if calls != want {
					t.Fatalf("preview fetched after authority change: %d", calls)
				}
			})
		}
	}
}
