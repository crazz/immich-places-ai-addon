package main

import (
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/writepreview"
)

func registerAIWritePreviewRoutes(mux *http.ServeMux, store *aiDraftStore, images *aiImagePreparer) {
	mux.HandleFunc("POST /ai/write-previews", func(w http.ResponseWriter, r *http.Request) {
		if !guardAIDraftMutation(w, r, store.results.origin) {
			return
		}
		var body struct {
			DraftID       string `json:"draftId"`
			DraftRevision int    `json:"draftRevision"`
		}
		media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		raw, readErr := io.ReadAll(http.MaxBytesReader(w, r.Body, 4<<10))
		if media != "application/json" || readErr != nil || jobs.DecodeRequest(raw, &body) != nil || body.DraftRevision < 1 || r.URL.RawQuery != "" {
			writeAIWritePreviewFailure(w, writepreview.Failure{Code: "INVALID_PREVIEW"})
			return
		}
		if _, err := uuid.Parse(body.DraftID); err != nil {
			writeAIWritePreviewFailure(w, writepreview.Failure{Code: "INVALID_PREVIEW"})
			return
		}
		session := &aiWritePreviewSession{drafts: store, images: images}
		value, err := writepreview.Create(r.Context(), session, session, session, getUserFromContext(r).ID, body.DraftID, body.DraftRevision, uuid.NewString(), store.results.jobs.now)
		if err != nil {
			writeAIWritePreviewFailure(w, err)
			return
		}
		writeJSON(w, 200, value)
	})
	mux.HandleFunc("GET /ai/write-previews/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			writeAIWritePreviewFailure(w, writepreview.Failure{Code: "INVALID_PREVIEW"})
			return
		}
		session := &aiWritePreviewSession{drafts: store}
		value, err := session.get(r.Context(), getUserFromContext(r).ID, r.PathValue("id"))
		if err != nil {
			writeAIWritePreviewFailure(w, err)
			return
		}
		writeJSON(w, 200, value)
	})
}

func writeAIWritePreviewFailure(w http.ResponseWriter, err error) {
	status, code := 503, "PREVIEW_UNAVAILABLE"
	var failure writepreview.Failure
	if errors.As(err, &failure) {
		code, status = failure.Code, 409
		if code == "INVALID_PREVIEW" {
			status = 400
		}
		if code == "SOURCE_UNAVAILABLE" {
			status = 503
		}
	} else if errors.Is(err, drafts.ErrUnavailable) {
		status = 404
	}
	if failure.Conflict != nil {
		writeJSON(w, status, struct {
			Code      string                 `json:"code"`
			Message   string                 `json:"message"`
			Retryable bool                   `json:"retryable"`
			RequestID string                 `json:"requestID"`
			Conflict  *writepreview.Conflict `json:"conflict"`
		}{code, "GPS changed. Explicitly review the current source and baseline, then stage a new revision.", false, uuid.NewString(), failure.Conflict})
		return
	}
	writeAIProviderError(w, status, code, "Preview unavailable or changed. Review the saved draft and current source before retrying.")
}
