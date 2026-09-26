package main

import (
	"context"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/translations"
	"net/http"
	"strings"
	"testing"
)

func TestAITranslationAdoptionRespectsConfirmedWriteRevision(t *testing.T) {
	for _, acting := range []bool{false, true} {
		t.Run(map[bool]string{false: "undispatched", true: "acting"}[acting], func(t *testing.T) {
			w := newAIWriteFixture(t)
			ctx := context.Background()
			s := &aiTranslationStore{drafts: w.writer.drafts, fingerprint: strings.Repeat("a", 64)}
			req := translations.Request{Key: "translate", DraftID: w.draft.ID, Revision: w.draft.Revision, FactsRevision: w.draft.FactsRevision, ProfileID: "profile", ProfileRevision: 1, Basis: "Reviewed bridge.", BasisKind: "candidate", Languages: []string{"en"}, Confirmed: true}
			run, err := s.submit(ctx, testUserID, req)
			if err != nil {
				t.Fatal(err)
			}
			selectionSQL(t, w.f.db, "UPDATE ai_translation_items SET state='complete',text='Reviewed bridge.' WHERE runID=?", run.ID)
			if acting {
				entered, release := make(chan struct{}), make(chan struct{})
				done := make(chan error, 1)
				w.handle = func(out http.ResponseWriter, r *http.Request) bool {
					if r.Method == "PATCH" {
						close(entered)
						<-release
					}
					return false
				}
				go func() { done <- w.writer.runOne(ctx, testUserID, w.op.ID) }()
				<-entered
				_, adoptErr := s.adopt(ctx, testUserID, run.ID, w.draft.Revision, []string{"en"})
				close(release)
				sendErr := <-done
				if adoptErr != drafts.ErrWriteInProgress || sendErr != nil {
					t.Fatal("acting revision not protected", adoptErr, sendErr)
				}
			} else {
				next, err := s.adopt(ctx, testUserID, run.ID, w.draft.Revision, []string{"en"})
				if err != nil || next.State != "draft" {
					t.Fatal("safe adoption rejected", next, err)
				}
				w.run(t)
				if sends, _ := w.counts(); sends != 0 {
					t.Fatal("obsolete approval sent", sends)
				}
			}
		})
	}
}
