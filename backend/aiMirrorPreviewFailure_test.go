package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIMirrorCombinedPreviewFailureRetainsLocalContentAndCreatesNoPlan(t *testing.T) {
	for _, kind := range []string{"stale", "oversized", "unavailable", "foreign"} {
		t.Run(kind, func(t *testing.T) {
			f := mirrorWriteFixture(t)
			ctx := context.Background()
			d := f.draft
			switch kind {
			case "stale":
				d.Descriptions[0].Stale = true
			case "oversized":
				d.Descriptions = nil
				d.Mirror = &drafts.MirrorSelection{}
				for _, tag := range []string{"en", "fr", "de", "es", "pt", "it", "nl", "ja"} {
					text := strings.Repeat("x", 9<<10)
					d.Descriptions = append(d.Descriptions, drafts.Description{Language: tag, Status: "complete", Text: &text, FactsRevision: d.FactsRevision})
					d.Mirror.Languages = append(d.Mirror.Languages, tag)
				}
			case "unavailable":
				f.handle = func(w http.ResponseWriter, r *http.Request) bool {
					if strings.HasSuffix(r.URL.Path, "/metadata") {
						w.WriteHeader(400)
						return true
					}
					return false
				}
			case "foreign":
				f.namespace = json.RawMessage(`{"schemaVersion":1,"application":"foreign"}`)
			}
			if kind == "stale" || kind == "oversized" {
				d.Revision++
				err := f.writer.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
					if err := f.writer.drafts.snapshot(ctx, tx, testUserID, d); err != nil {
						return err
					}
					_, err := tx.ExecContext(ctx, `UPDATE ai_drafts SET revision=? WHERE userID=? AND id=?`, d.Revision, testUserID, d.ID)
					return err
				})
				if err != nil {
					t.Fatal(err)
				}
			}
			before, _ := json.Marshal(d)
			var count int
			if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_stack_write_previews`).Scan(&count); err != nil {
				t.Fatal(err)
			}
			f.writer.drafts.results.origin = aiTestOrigin
			raw := fmt.Sprintf(`{"draftId":%q,"draftRevision":%d,"mirrorDisclosure":%q}`, d.ID, d.Revision, writepreview.MirrorDisclosure)
			rec := aiRequest(newAIResultHandler(f.writer.drafts.results, f.image.service), "POST", "/ai/write-previews", raw, aiTestOrigin, true)
			if rec.Code == 200 {
				t.Fatal("invalid combined preview was usable")
			}
			var afterCount int
			if err := f.f.db.db.QueryRow(`SELECT count(*) FROM ai_stack_write_previews`).Scan(&afterCount); err != nil || count != afterCount {
				t.Fatal("failed preview persisted plan", err)
			}
			stored, err := f.writer.drafts.get(ctx, testUserID, d.ID)
			after, _ := json.Marshal(stored)
			if err != nil || string(before) != string(after) {
				t.Fatal("failure discarded local reviewed data", err)
			}
			if _, bad := f.image.counts(); bad != 0 {
				t.Fatal("failed preview mutated")
			}
		})
	}
}

func TestAIMirrorPreviewReloadIsStrictlyLocal(t *testing.T) {
	f := mirrorWriteFixture(t)
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		t.Error("reload issued upstream request")
		w.WriteHeader(503)
		return true
	}
	before, bad := f.image.counts()
	s := &aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service}
	got, err := s.get(context.Background(), testUserID, f.preview.Plan.ID)
	after, later := f.image.counts()
	if err != nil || got.Digest != f.preview.Digest || before != after || bad != later {
		t.Fatal("reload was not exact local read", err)
	}
	if _, err := s.get(context.Background(), uuid.NewString(), f.preview.Plan.ID); err == nil {
		t.Fatal("foreign read")
	}
}
