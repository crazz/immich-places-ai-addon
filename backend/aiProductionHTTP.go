package main

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"

	"immich-places-backend/internal/ai/jobs"
)

func newAIJobHandler(p *aiProductionJobs, origin string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ai/jobs", func(w http.ResponseWriter, r *http.Request) {
		query, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			writeAIJobFailure(w, jobs.ErrInvalid)
			return
		}
		for key, values := range query {
			if (key != "limit" && key != "cursor") || len(values) != 1 {
				writeAIJobFailure(w, jobs.ErrInvalid)
				return
			}
		}
		limit := 20
		if value, ok := query["limit"]; ok {
			limit, err = strconv.Atoi(value[0])
			if err != nil {
				writeAIJobFailure(w, jobs.ErrInvalid)
				return
			}
		}
		result, err := p.list(r.Context(), getUserFromContext(r).ID, limit, query.Get("cursor"))
		if err != nil {
			writeAIJobFailure(w, err)
			return
		}
		writeJSON(w, 200, result)
	})
	mux.HandleFunc("POST /ai/jobs", func(w http.ResponseWriter, r *http.Request) {
		if !p.store.enabled {
			writeAIProviderError(w, 503, "AI_DISABLED", "AI is disabled")
			return
		}
		var req jobs.Admission
		if !decodeAIJobBody(w, r, &req) {
			return
		}
		job, err := p.submit(r.Context(), getUserFromContext(r).ID, req)
		if err != nil {
			writeAIJobFailure(w, err)
			return
		}
		result, err := p.progress(r.Context(), getUserFromContext(r).ID, job.ID)
		if err != nil {
			writeAIJobFailure(w, err)
			return
		}
		writeJSON(w, 202, result)
	})
	mux.HandleFunc("GET /ai/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		result, err := p.progress(r.Context(), getUserFromContext(r).ID, r.PathValue("id"))
		if err != nil {
			writeAIJobFailure(w, err)
			return
		}
		writeJSON(w, 200, result)
	})
	mux.HandleFunc("POST /ai/jobs/{id}/cancel", func(w http.ResponseWriter, r *http.Request) {
		var body struct{}
		if !decodeAIJobBody(w, r, &body) {
			return
		}
		owner := getUserFromContext(r).ID
		if _, err := p.progress(r.Context(), owner, r.PathValue("id")); err != nil {
			writeAIJobFailure(w, err)
			return
		}
		if err := p.store.Cancel(r.Context(), owner, r.PathValue("id")); err != nil {
			writeAIJobFailure(w, err)
			return
		}
		writeJSON(w, 200, struct {
			Canceled bool `json:"canceled"`
		}{true})
	})
	guarded := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && (len(r.Header.Values("Origin")) != 1 || origin == "" || r.Header.Get("Origin") != origin) {
			writeAIProviderError(w, 403, "ORIGIN_REJECTED", "request origin is not allowed")
			return
		}
		mux.ServeHTTP(w, r)
	})
	protected := sessionMiddlewareWithErrors(p.store.db, guarded, func(w http.ResponseWriter, status int, message string) {
		code := "UNAUTHENTICATED"
		if status == http.StatusInternalServerError {
			code = "STORAGE_ERROR"
		}
		writeAIProviderError(w, status, code, message)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		protected.ServeHTTP(w, r)
	})
}
func decodeAIJobBody(w http.ResponseWriter, r *http.Request, target any) bool {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		writeAIProviderError(w, 415, "INVALID_REQUEST", "Content-Type must be application/json")
		return false
	}

	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 32<<10))
	if err != nil || jobs.DecodeRequest(data, target) != nil {
		writeAIJobFailure(w, jobs.ErrInvalid)
		return false
	}

	return true
}
func writeAIJobFailure(w http.ResponseWriter, err error) {
	status, code, message := 503, "STORAGE_ERROR", "job storage is unavailable; retry with the same submission key"
	switch {
	case errors.Is(err, jobs.ErrInvalid):
		status, code, message = 400, "INVALID_JOB", "provide a valid exact configuration, consent and finite limits"
	case errors.Is(err, jobs.ErrDenied):
		status, code, message = 404, "JOB_UNAVAILABLE", "job or required authority is unavailable"
	case errors.Is(err, jobs.ErrConflict):
		status, code, message = 409, "JOB_CONFLICT", "submission key already belongs to different input"
	case errors.Is(err, jobs.ErrBudget):
		status, code, message = 409, "JOB_BUDGET", "job resource allowance is exhausted"
	}
	writeAIProviderError(w, status, code, message)
}
