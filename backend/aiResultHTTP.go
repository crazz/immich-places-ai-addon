package main

import (
	"errors"
	"net/http"
	"net/url"

	"immich-places-backend/internal/ai/review"
)

func newAIResultHandler(s *aiResultStore, images *aiImagePreparer) http.Handler {
	mux := http.NewServeMux()
	registerAIDraftRoutes(mux, s, images)
	mux.HandleFunc("GET /ai/results", func(w http.ResponseWriter, r *http.Request) {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil || len(r.URL.RawQuery) > 4096 {
			writeAIResultFailure(w, review.ErrInvalid, false)
			return
		}
		q, err := review.ParseQuery(values)
		if err != nil {
			writeAIResultFailure(w, err, false)
			return
		}
		page, err := s.list(r.Context(), getUserFromContext(r).ID, q)
		if err != nil {
			writeAIResultFailure(w, err, false)
			return
		}
		writeJSON(w, 200, page)
	})
	detail := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			writeAIResultFailure(w, review.ErrInvalid, true)
			return
		}
		result, err := s.detail(r.Context(), getUserFromContext(r).ID, r.PathValue("analysisID"), r.PathValue("jobID"), r.PathValue("itemID"))
		if err != nil {
			writeAIResultFailure(w, err, true)
			return
		}
		writeJSON(w, 200, result)
	}
	mux.HandleFunc("GET /ai/results/{analysisID}", detail)
	mux.HandleFunc("GET /ai/jobs/{jobID}/items/{itemID}/result", detail)
	mux.HandleFunc("GET /ai/jobs/{jobID}/items/{itemID}/thumbnail", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			writeAIResultFailure(w, review.ErrInvalid, true)
			return
		}
		data, err := s.thumbnail(r.Context(), getUserFromContext(r).ID, r.PathValue("jobID"), r.PathValue("itemID"), images)
		if err != nil {
			writeAIResultFailure(w, err, true)
			return
		}
		defer clear(data)
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_, _ = w.Write(data)
	})
	protected := sessionMiddlewareWithErrors(s.jobs.db, mux, func(w http.ResponseWriter, status int, _ string) {
		code := "UNAUTHENTICATED"
		if status == http.StatusInternalServerError {
			code = "STORAGE_ERROR"
		}
		writeAIProviderError(w, status, code, "result access unavailable")
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		protected.ServeHTTP(w, r)
	})
}

func writeAIResultFailure(w http.ResponseWriter, err error, detail bool) {
	if errors.Is(err, review.ErrInvalid) {
		writeAIProviderError(w, 400, "INVALID_RESULT_QUERY", "provide valid history filters and continuation")
		return
	}
	if detail {
		writeAIProviderError(w, 404, "RESULT_UNAVAILABLE", "result unavailable")
		return
	}
	writeAIProviderError(w, 503, "HISTORY_UNAVAILABLE", "history is temporarily unavailable; retry explicitly")
}
