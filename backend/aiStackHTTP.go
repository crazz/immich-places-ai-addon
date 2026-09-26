package main

import (
	"io"
	"mime"
	"net/http"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/writepreview"
)

func registerAIStackReviewRoutes(mux *http.ServeMux, store *aiDraftStore, images *aiImagePreparer) {
	mux.HandleFunc("POST /ai/stack-reviews", func(w http.ResponseWriter, r *http.Request) {
		if !guardAIDraftMutation(w, r, store.results.origin) {
			return
		}
		var body struct {
			DraftID       string   `json:"draftId"`
			DraftRevision int      `json:"draftRevision"`
			TargetIDs     []string `json:"targetIds"`
		}
		media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 8<<10))
		if err != nil || media != "application/json" || r.URL.RawQuery != "" || jobs.DecodeRequest(raw, &body) != nil || body.DraftRevision < 1 {
			writeAIWritePreviewFailure(w, writepreview.Failure{Code: "INVALID_PREVIEW"})
			return
		}
		if _, err = uuid.Parse(body.DraftID); err != nil {
			writeAIWritePreviewFailure(w, writepreview.Failure{Code: "INVALID_PREVIEW"})
			return
		}
		seen := make(map[string]bool)
		for _, id := range body.TargetIDs {
			parsed, err := uuid.Parse(id)
			if err != nil || parsed.String() != id {
				writeAIWritePreviewFailure(w, writepreview.Failure{Code: "INVALID_PREVIEW"})
				return
			}
			seen[id] = true
		}
		if len(seen) > writepreview.MaxStackTargets {
			writeAIWritePreviewFailure(w, writepreview.Failure{Code: "INVALID_PREVIEW"})
			return
		}
		review, err := store.observeStack(r.Context(), getUserFromContext(r).ID, body.DraftID, body.DraftRevision, body.TargetIDs, images)
		if err != nil {
			writeAIWritePreviewFailure(w, err)
			return
		}
		writeJSON(w, 200, review)
	})
}
