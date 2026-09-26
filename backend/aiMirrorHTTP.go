package main

import (
	"net/http"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/writeback"
)

func registerAIMirrorActions(mux *http.ServeMux, store *aiWriteStore, origin string) {
	for _, action := range []string{"retry", "reconcile"} {
		mux.HandleFunc("POST /ai/write-operations/{id}/targets/{assetId}/metadata/"+action, func(w http.ResponseWriter, r *http.Request) {
			if !guardAIDraftMutation(w, r, origin) {
				return
			}
			var input struct {
				Generation *int `json:"generation"`
			}
			if !decodeAIWriteBody(w, r, &input) {
				return
			}
			id, asset := r.PathValue("id"), r.PathValue("assetId")
			parsedID, idErr := uuid.Parse(id)
			parsedAsset, assetErr := uuid.Parse(asset)
			if idErr != nil || assetErr != nil || parsedID.String() != id || parsedAsset.String() != asset || input.Generation == nil || *input.Generation < 1 {
				writeAIWriteFailure(w, writeback.Failure("INVALID_WRITE"))
				return
			}
			var op writeback.Operation
			var err error
			if action == "retry" {
				op, err = store.retryMirror(r.Context(), getUserFromContext(r).ID, id, asset, *input.Generation)
			} else {
				op, err = store.reconcileMirror(r.Context(), getUserFromContext(r).ID, id, asset, *input.Generation)
			}
			if err != nil {
				writeAIWriteFailure(w, err)
				return
			}
			writeJSON(w, 200, op)
		})
	}
}
