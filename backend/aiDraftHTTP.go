package main

import (
	"errors"
	"net/http"
	"strconv"

	"immich-places-backend/internal/ai/drafts"
)

func registerAIDraftRoutes(mux *http.ServeMux, results *aiResultStore, images *aiImagePreparer) {
	store := &aiDraftStore{results: results}
	registerAIWritePreviewRoutes(mux, store, images)
	registerAIDraftBaselineRoutes(mux, store, images)
	mux.HandleFunc("POST /ai/results/{analysisID}/draft", func(w http.ResponseWriter, r *http.Request) {
		if !guardAIDraftMutation(w, r, results.origin) {
			return
		}
		var body struct {
			CandidateID *string `json:"candidateId"`
		}
		if !decodeAIJobBody(w, r, &body) {
			return
		}
		value, err := store.accept(r.Context(), getUserFromContext(r).ID, r.PathValue("analysisID"), body.CandidateID)
		if err != nil {
			writeAIDraftFailure(w, err)
			return
		}
		writeJSON(w, 200, value)
	})
	mux.HandleFunc("PATCH /ai/drafts/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !guardAIDraftMutation(w, r, results.origin) {
			return
		}
		revision, ok := aiDraftRevision(w, r)
		if !ok {
			return
		}
		var body drafts.Edit
		if !decodeAIJobBody(w, r, &body) {
			return
		}
		value, err := store.edit(r.Context(), getUserFromContext(r).ID, r.PathValue("id"), revision, body)
		if err != nil {
			writeAIDraftFailure(w, err)
			return
		}
		writeJSON(w, 200, value)
	})

	mux.HandleFunc("GET /ai/drafts/{id}", func(w http.ResponseWriter, r *http.Request) {
		value, err := store.get(r.Context(), getUserFromContext(r).ID, r.PathValue("id"))
		if err != nil {
			writeAIDraftFailure(w, err)
			return
		}
		writeJSON(w, 200, value)
	})
}

func writeAIDraftFailure(w http.ResponseWriter, err error) {
	status, code := 503, "STORAGE_ERROR"
	switch {
	case errors.Is(err, drafts.ErrWriteInProgress):
		status, code = 409, "WRITE_IN_PROGRESS"
	case errors.Is(err, drafts.ErrUnavailable):
		status, code = 404, "DRAFT_UNAVAILABLE"
	case errors.Is(err, drafts.ErrInvalid):
		status, code = 400, "INVALID_DRAFT"
	case errors.Is(err, drafts.ErrConflict):
		status, code = 412, "DRAFT_CONFLICT"
	}
	writeAIProviderError(w, status, code, "draft unavailable or changed; reload saved state before retrying")
}

func guardAIDraftMutation(w http.ResponseWriter, r *http.Request, origin string) bool {
	if len(r.Header.Values("Origin")) != 1 || origin == "" || r.Header.Get("Origin") != origin {
		writeAIProviderError(w, 403, "ORIGIN_REJECTED", "request origin is not allowed")
		return false
	}
	return true
}
func aiDraftRevision(w http.ResponseWriter, r *http.Request) (int, bool) {
	token := r.Header.Get("If-Match")
	if token == "" {
		writeAIProviderError(w, 428, "REVISION_REQUIRED", "reload the saved draft and provide its revision")
		return 0, false
	}
	raw, err := strconv.Unquote(token)
	revision, parseErr := strconv.Atoi(raw)
	if err != nil || parseErr != nil || revision < 1 || len(r.Header.Values("If-Match")) != 1 || token != strconv.Quote(strconv.Itoa(revision)) {
		writeAIDraftFailure(w, drafts.ErrInvalid)
		return 0, false
	}
	return revision, true
}
