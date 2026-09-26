package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestAITranslationDeletionAndRotationFenceActiveSender(t *testing.T) {
	for _, change := range []string{"delete", "rotate"} {
		t.Run(change, func(t *testing.T) {
			f, s, req := translationFixture(t)
			entered, release := make(chan struct{}), make(chan struct{})
			dispatcher := translationProvider(t, f, s, func(w http.ResponseWriter, r *http.Request) {
				close(entered)
				<-release
				_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": `{"language":"en","status":"complete","text":"Must not publish."}`}}}})
			})
			req.ProfileRevision = 2
			run, err := s.submit(context.Background(), testUserID, req)
			if err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { _, err := s.runOne(context.Background(), dispatcher); done <- err }()
			<-entered
			if change == "delete" {
				_, err = f.db.db.Exec("DELETE FROM users WHERE id=?", testUserID)
			} else {
				err = f.selection.bind(context.Background(), "https://replacement.example/api", "translation-epoch")
			}
			close(release)
			<-done
			if err != nil {
				t.Fatal(err)
			}
			var active, texts, runs int
			if err = f.db.db.QueryRow("SELECT count(*) FROM ai_translation_items WHERE state IN ('queued','reserved')").Scan(&active); err != nil || active != 0 {
				t.Fatal("private work survived lifecycle change", active, err)
			}
			if err = f.db.db.QueryRow("SELECT count(*) FROM ai_translation_items WHERE text IS NOT NULL").Scan(&texts); err != nil || texts != 0 {
				t.Fatal("late text published", texts, err)
			}
			if err = f.db.db.QueryRow("SELECT count(*) FROM ai_translation_runs WHERE id=?", run.ID).Scan(&runs); err != nil || (change == "delete" && runs != 0) || (change == "rotate" && runs != 1) {
				t.Fatal("incorrect history retention", runs, err)
			}
		})
	}
}
