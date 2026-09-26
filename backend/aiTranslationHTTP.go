package main

import (
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/translations"
	"net/http"
	"net/url"
)

func newAITranslationHandler(s *aiTranslationStore, origin string) http.Handler {
	mux := http.NewServeMux()
	registerAITranslationActions(mux, s, origin)
	mux.HandleFunc("GET /ai/translations", func(w http.ResponseWriter, r *http.Request) {
		q, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil || len(r.URL.RawQuery) > 512 || q.Get("draftId") == "" {
			writeAIDraftFailure(w, drafts.ErrInvalid)
			return
		}
		for k, v := range q {
			if (k != "draftId" && k != "before") || len(v) != 1 {
				writeAIDraftFailure(w, drafts.ErrInvalid)
				return
			}
		}
		page, err := s.list(r.Context(), getUserFromContext(r).ID, q.Get("draftId"), q.Get("before"))
		if err != nil {
			writeAIDraftFailure(w, err)
			return
		}
		writeJSON(w, 200, page)
	})
	mux.HandleFunc("POST /ai/translations", func(w http.ResponseWriter, r *http.Request) {
		if !guardAIDraftMutation(w, r, origin) {
			return
		}
		var req translations.Request
		if !decodeAIJobBody(w, r, &req) {
			return
		}
		run, err := s.submit(r.Context(), getUserFromContext(r).ID, req)
		if err != nil {
			writeAIDraftFailure(w, err)
			return
		}
		writeJSON(w, 200, run)
	})
	mux.HandleFunc("GET /ai/translations/{id}", func(w http.ResponseWriter, r *http.Request) {
		run, err := s.get(r.Context(), getUserFromContext(r).ID, r.PathValue("id"))
		if err != nil {
			writeAIDraftFailure(w, err)
			return
		}
		writeJSON(w, 200, run)
	})
	protected := sessionMiddlewareWithErrors(s.drafts.results.jobs.db, mux, func(w http.ResponseWriter, status int, _ string) {
		writeAIProviderError(w, status, "TRANSLATION_UNAVAILABLE", "translation access unavailable")
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		protected.ServeHTTP(w, r)
	})
}
