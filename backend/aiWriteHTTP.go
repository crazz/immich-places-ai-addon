package main

import (
	"errors"
	"io"
	"mime"
	"net/http"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/writeback"
)

func registerAIWriteRoutes(mux *http.ServeMux, results *aiResultStore) {
	store := results.writer
	if store == nil {
		store = &aiWriteStore{drafts: &aiDraftStore{results: results}}
	}
	registerAIWriteRetryRoute(mux, store, results.origin)
	registerAIWriteReconcileRoute(mux, store, results.origin)
	registerAIWriteHistoryRoute(mux, store)
	mux.HandleFunc("POST /ai/write-operations", func(w http.ResponseWriter, r *http.Request) {
		if !guardAIDraftMutation(w, r, results.origin) {
			return
		}
		var body writeback.Confirmation
		if !decodeAIWriteBody(w, r, &body) {
			return
		}
		op, err := store.confirm(r.Context(), getUserFromContext(r).ID, body)
		if err != nil {
			writeAIWriteFailure(w, err)
			return
		}
		writeJSON(w, 200, op)
	})
	for _, route := range []string{"GET /ai/write-operations/{id}", "GET /ai/write-operations/by-key/{key}"} {
		mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.RawQuery != "" {
				writeAIWriteFailure(w, writeback.Failure("INVALID_WRITE"))
				return
			}
			id, byKey := r.PathValue("id"), false
			if id == "" {
				id, byKey = r.PathValue("key"), true
			}
			op, err := store.get(r.Context(), getUserFromContext(r).ID, id, byKey)
			if err != nil {
				writeAIWriteFailure(w, err)
				return
			}
			writeJSON(w, 200, op)
		})
	}
}

func decodeAIWriteBody(w http.ResponseWriter, r *http.Request, value any) bool {
	media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 4<<10))
	if media != "application/json" || err != nil || r.URL.RawQuery != "" || jobs.DecodeRequest(raw, value) != nil {
		writeAIWriteFailure(w, writeback.Failure("INVALID_WRITE"))
		return false
	}
	return true
}

func writeAIWriteFailure(w http.ResponseWriter, err error) {
	code, status := "WRITE_UNAVAILABLE", 503
	var failure writeback.Failure
	if errors.Is(err, drafts.ErrUnavailable) {
		status = 404
	}
	if errors.As(err, &failure) {
		code = string(failure)
		status = 409
		switch code {
		case "INVALID_WRITE":
			status = 400
		case "PREVIEW_EXPIRED":
			status = 410
		case "WRITE_DISABLED", "WRITE_UNAVAILABLE":
			status = 503
		}
	}
	writeAIProviderError(w, status, code, "GPS operation unavailable or changed; inspect saved state before continuing")
}
