package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestAITranslationFencesChangedAuthorityBeforeDispatchAndPublication(t *testing.T) {
	for _, at := range []string{"before", "resolved", "publication"} {
		t.Run(at, func(t *testing.T) {
			f, s, req := translationFixture(t)
			var hits atomic.Int32
			dispatcher := translationProvider(t, f, s, func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				if at == "publication" {
					if _, err := f.db.db.Exec("UPDATE ai_provider_profiles SET enabled=0 WHERE id='profile'"); err != nil {
						t.Error(err)
					}
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": `{"language":"en","status":"complete","text":"Unapproved late result."}`}}}})
			})
			req.ProfileRevision = 2
			req.Languages = []string{"en"}
			ctx := context.Background()
			if _, err := s.submit(ctx, testUserID, req); err != nil {
				t.Fatal(err)
			}
			if at == "before" {
				s.fingerprint = strings.Repeat("b", 64)
			}
			if at == "resolved" {
				dispatcher.AfterResolve = func(context.Context) { s.fingerprint = strings.Repeat("b", 64) }
			}
			if _, err := s.runOne(ctx, dispatcher); err != nil {
				t.Fatal(err)
			}
			run, err := s.submit(ctx, testUserID, req)
			if err != nil || run.Items[0].Text != nil || run.Items[0].State == "complete" {
				t.Fatal("stale authority published output", run, err)
			}
			if (at == "publication" && hits.Load() != 1) || (at != "publication" && hits.Load() != 0) {
				t.Fatal("obsolete authority dispatched", hits.Load())
			}
		})
	}
}
