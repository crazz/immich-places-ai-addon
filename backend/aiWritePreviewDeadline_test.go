package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAIWritePreviewBoundsSourceDeadlineAndCancellation(t *testing.T) {
	for _, cancelEarly := range []bool{false, true} {
		t.Run(map[bool]string{false: "deadline", true: "cancel"}[cancelEarly], func(t *testing.T) {
			f, image, store, draft, _ := writePreviewFixture(t)
			entered := make(chan struct{})
			image.handle = func(w http.ResponseWriter, r *http.Request) bool {
				close(entered)
				<-r.Context().Done()
				return true
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			body, _ := json.Marshal(map[string]any{"draftId": draft.ID, "draftRevision": draft.Revision})
			req := httptest.NewRequest("POST", "/ai/write-previews", strings.NewReader(string(body))).WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", aiTestOrigin)
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "ai-session"})
			rec := httptest.NewRecorder()
			done := make(chan struct{})
			reads, bad := image.counts()
			start := time.Now()
			go func() { writePreviewHandler(f, image).ServeHTTP(rec, req); close(done) }()
			<-entered
			if cancelEarly {
				cancel()
			}
			<-done
			if rec.Code != 503 || !strings.Contains(rec.Body.String(), "SOURCE_UNAVAILABLE") || time.Since(start) >= 12*time.Second {
				t.Fatal("unbounded or unsafe source failure", rec.Code, time.Since(start), rec.Body.String())
			}
			var count int
			if err := f.db.db.QueryRow("SELECT count(*) FROM ai_write_previews").Scan(&count); err != nil || count != 0 {
				t.Fatal("failed source published", count, err)
			}
			retained, err := store.get(context.Background(), testUserID, draft.ID)
			if err != nil || retained.Revision != draft.Revision || retained.State != draft.State || *retained.Camera != *draft.Camera {
				t.Fatal("failed source changed draft", err)
			}
			if after, failures := image.counts(); after != reads+1 || failures != bad {
				t.Fatal("failed source retried or mutated", after, failures)
			}
		})
	}
}
