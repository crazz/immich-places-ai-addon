package main

import (
	"context"
	"database/sql"
	"net/http"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func registerAIWriteRetryRoute(mux *http.ServeMux, store *aiWriteStore, origin string) {
	mux.HandleFunc("POST /ai/write-operations/{id}/retry", func(w http.ResponseWriter, r *http.Request) {
		if !guardAIDraftMutation(w, r, origin) {
			return
		}
		var input struct {
			Generation *int `json:"generation"`
		}
		if !decodeAIWriteBody(w, r, &input) {
			return
		}
		if input.Generation == nil || *input.Generation < 1 {
			writeAIWriteFailure(w, writeback.Failure("INVALID_WRITE"))
			return
		}
		op, err := store.retry(r.Context(), getUserFromContext(r).ID, r.PathValue("id"), *input.Generation)
		if err != nil {
			writeAIWriteFailure(w, err)
			return
		}
		writeJSON(w, 200, op)
	})
}

func (s *aiWriteStore) retry(ctx context.Context, owner, id string, generation int) (writeback.Operation, error) {
	op, err := s.get(ctx, owner, id, false)
	if err != nil {
		return op, err
	}
	a := &aiWriteAttempt{store: s}
	fresh, err := a.Read(ctx, op)
	if err != nil {
		return op, writeback.Failure("RETRY_UNAVAILABLE")
	}
	if fresh.ImageIdentity != op.Plan.ImageIdentity || !writepreview.EqualGPS(fresh.GPS, op.Plan.Before) {
		return op, writeback.Failure("RETRY_UNAVAILABLE")
	}
	err = s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, err := s.read(ctx, tx, owner, id, false)
		if err != nil {
			return err
		}
		var from int
		var known, active bool
		if tx.QueryRowContext(ctx, `SELECT retryFrom,completionKnown,senderActive FROM ai_write_targets WHERE userID=? AND installationID=? AND operationID=?`, owner, s.drafts.results.jobs.binding, id).Scan(&from, &known, &active) != nil {
			return drafts.ErrStorage
		}
		op = current
		if from == generation {
			return nil
		}
		if current.Status != "retryable" || current.Generation != generation || !known || active || current.Attempts >= 2 {
			return writeback.Failure("RETRY_UNAVAILABLE")
		}
		if err = s.dispatchAuthority(ctx, tx, current, a.authority); err != nil {
			return err
		}
		var guard string
		if tx.QueryRowContext(ctx, `SELECT token FROM ai_write_target_guards WHERE installationID=? AND assetID=?`, op.Plan.Installation, op.Plan.TargetID).Scan(&guard) != nil || guard != id {
			return writeback.Failure("TARGET_BUSY")
		}
		if _, err = tx.ExecContext(ctx, `UPDATE ai_write_targets SET retryFrom=?,generation=generation+1,reads=0,dueAt=0 WHERE userID=? AND installationID=? AND operationID=?`, generation, owner, op.Plan.Installation, id); err != nil {
			return drafts.ErrStorage
		}
		if _, err = tx.ExecContext(ctx, `UPDATE ai_write_operations SET status='queued',code='EXPLICIT_RETRY' WHERE userID=? AND installationID=? AND id=?`, owner, op.Plan.Installation, id); err != nil {
			return drafts.ErrStorage
		}
		if err = s.event(ctx, tx, current, "EXPLICIT_RETRY"); err != nil {
			return err
		}
		op, err = s.read(ctx, tx, owner, id, false)
		return err
	})
	return op, err
}
