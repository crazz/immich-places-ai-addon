package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func registerAIStackTargetActions(mux *http.ServeMux, store *aiWriteStore, origin string) {
	for _, action := range []string{"retry", "reconcile"} {
		mux.HandleFunc("POST /ai/write-operations/{id}/targets/{assetId}/"+action, func(w http.ResponseWriter, r *http.Request) {
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
				op, err = store.retryStackTarget(r.Context(), getUserFromContext(r).ID, id, asset, *input.Generation)
			} else {
				op, err = store.reconcileStackTarget(r.Context(), getUserFromContext(r).ID, id, asset, *input.Generation)
			}
			if err != nil {
				writeAIWriteFailure(w, err)
				return
			}
			writeJSON(w, 200, op)
		})
	}
}

func (s *aiWriteStore) reconcileStackTarget(ctx context.Context, owner, id, asset string, generation int) (writeback.Operation, error) {
	var parent writeback.Operation
	err := s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		parent, err = s.readStackOperation(ctx, tx, owner, id, false)
		if err != nil {
			return drafts.ErrUnavailable
		}
		step, err := writeback.TargetStep(parent, asset)
		if err != nil || generation < 1 || step.Generation != generation || (step.Status != "writing" && step.Status != "verifying") {
			return writeback.Failure("RECONCILIATION_UNAVAILABLE")
		}
		now := s.drafts.results.jobs.now()
		result, err := tx.ExecContext(ctx, `UPDATE ai_stack_write_targets SET reads=0,dueAt=?,generation=generation+1,leaseToken=?,leaseUntil=0 WHERE userID=? AND installationID=? AND operationID=? AND assetID=? AND generation=? AND reads=3 AND (senderActive=0 OR leaseUntil<=?)`, now.Add(time.Second).UnixNano(), uuid.NewString(), owner, parent.Plan.Installation, id, asset, generation, now.UnixNano())
		if err != nil {
			return drafts.ErrStorage
		}
		if n, err := result.RowsAffected(); err != nil || n != 1 {
			return writeback.Failure("RECONCILIATION_UNAVAILABLE")
		}
		a := &aiStackAttempt{store: s, parent: parent, assetID: asset}
		if err := a.event(ctx, tx, "EXPLICIT_RECONCILIATION", step.Attempts); err != nil {
			return err
		}
		parent, err = s.readStackOperation(ctx, tx, owner, id, false)
		return err
	})
	return parent, err
}
