package main

import "net/http"

func registerAIDraftBaselineRoutes(mux *http.ServeMux, store *aiDraftStore, reader *aiImagePreparer) {
	mux.HandleFunc("POST /ai/drafts/{id}/baseline", func(w http.ResponseWriter, r *http.Request) {
		if !guardAIDraftMutation(w, r, store.results.origin) {
			return
		}
		revision, ok := aiDraftRevision(w, r)
		if !ok {
			return
		}
		var body struct {
			ObservationID string `json:"observationId"`
		}
		if !decodeAIJobBody(w, r, &body) {
			return
		}
		owner, id := getUserFromContext(r).ID, r.PathValue("id")
		if body.ObservationID == "" {
			observation, err := store.observe(r.Context(), owner, id, revision, reader)
			if err != nil {
				writeAIDraftFailure(w, err)
				return
			}
			writeJSON(w, 200, observation)
		} else {
			value, err := store.acknowledge(r.Context(), owner, id, revision, body.ObservationID, reader)
			if err != nil {
				writeAIDraftFailure(w, err)
				return
			}
			writeJSON(w, 200, value)
		}
	})
}
