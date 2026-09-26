package main

import (
	"net/http"
)

func registerAITranslationActions(mux *http.ServeMux, s *aiTranslationStore, origin string) {
	mux.HandleFunc("POST /ai/translations/{id}/adopt", func(w http.ResponseWriter, r *http.Request) {
		if !guardAIDraftMutation(w, r, origin) {
			return
		}
		var body struct {
			Revision  int      `json:"revision"`
			Languages []string `json:"languages"`
		}
		if !decodeAIJobBody(w, r, &body) {
			return
		}
		value, err := s.adopt(r.Context(), getUserFromContext(r).ID, r.PathValue("id"), body.Revision, body.Languages)
		if err != nil {
			writeAIDraftFailure(w, err)
			return
		}
		writeJSON(w, 200, value)
	})
	mux.HandleFunc("POST /ai/translations/{id}/cancel", func(w http.ResponseWriter, r *http.Request) {
		if !guardAIDraftMutation(w, r, origin) {
			return
		}
		var body struct{}
		if !decodeAIJobBody(w, r, &body) {
			return
		}
		owner, id := getUserFromContext(r).ID, r.PathValue("id")
		if err := s.cancel(r.Context(), owner, id); err != nil {
			writeAIDraftFailure(w, err)
			return
		}
		run, err := s.get(r.Context(), owner, id)
		if err != nil {
			writeAIDraftFailure(w, err)
			return
		}
		writeJSON(w, 200, run)
	})
}
