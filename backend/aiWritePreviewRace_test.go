package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"immich-places-backend/internal/ai/drafts"
)

func TestAIWritePreviewFencesConcurrentDraftAuthorityAndStorageChanges(t *testing.T) {
	for _, change := range []string{"edit", "reject", "delete", "credential", "installation", "hidden", "storage"} {
		t.Run(change, func(t *testing.T) {
			f, image, store, draft, meta := writePreviewFixture(t)
			entered, release := make(chan struct{}), make(chan struct{})
			image.handle = func(w http.ResponseWriter, r *http.Request) bool {
				close(entered)
				<-release
				_ = json.NewEncoder(w).Encode(meta)
				return true
			}
			body, _ := json.Marshal(map[string]any{"draftId": draft.ID, "draftRevision": draft.Revision})
			done := make(chan int, 1)
			go func() {
				rec := aiRequest(writePreviewHandler(f, image), "POST", "/ai/write-previews", string(body), aiTestOrigin, true)
				done <- rec.Code
			}()
			<-entered
			ctx := context.Background()
			switch change {
			case "edit":
				if _, err := store.edit(ctx, testUserID, draft.ID, draft.Revision, drafts.Edit{Camera: json.RawMessage(`{"latitude":2,"longitude":3}`)}); err != nil {
					t.Fatal(err)
				}
			case "reject":
				if _, err := store.edit(ctx, testUserID, draft.ID, draft.Revision, drafts.Edit{State: "rejected"}); err != nil {
					t.Fatal(err)
				}
			case "delete":
				selectionSQL(t, f.db, "DELETE FROM users WHERE ID=?", testUserID)
			case "credential":
				key := "changed"
				if err := f.db.updateImmichAPIKey(ctx, testUserID, &key); err != nil {
					t.Fatal(err)
				}
			case "installation":
				selectionSQL(t, f.db, "UPDATE ai_installation_identity SET id=?", selectionB)
			case "hidden":
				selectionSQL(t, f.db, "UPDATE assets SET isHidden=1 WHERE userID=?", testUserID)
			case "storage":
				selectionSQL(t, f.db, `CREATE TRIGGER preview_fail BEFORE INSERT ON ai_write_previews BEGIN SELECT RAISE(ABORT,'private storage value'); END`)
			}
			close(release)
			if status := <-done; status == 200 {
				t.Fatal("obsolete authority published")
			}
			var count int
			if err := f.db.db.QueryRow("SELECT count(*) FROM ai_write_previews").Scan(&count); err != nil || count != 0 {
				t.Fatal("obsolete plan persisted", count, err)
			}
			if _, bad := image.counts(); bad != 0 {
				t.Fatal("unexpected mutation")
			}
		})
	}
}
