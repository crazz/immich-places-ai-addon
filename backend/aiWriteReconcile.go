package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func registerAIWriteReconcileRoute(mux *http.ServeMux, store *aiWriteStore, origin string) {
	mux.HandleFunc("POST /ai/write-operations/{id}/reconcile", func(w http.ResponseWriter, r *http.Request) {
		if !guardAIDraftMutation(w, r, origin) {
			return
		}
		var body struct{}
		if !decodeAIWriteBody(w, r, &body) {
			return
		}
		op, err := store.reconcile(r.Context(), getUserFromContext(r).ID, r.PathValue("id"))
		if err != nil {
			writeAIWriteFailure(w, err)
			return
		}
		writeJSON(w, 200, op)
	})
}

func (s *aiWriteStore) reconcile(ctx context.Context, owner, id string) (writeback.Operation, error) {
	var op writeback.Operation
	err := s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		op, err = s.read(ctx, tx, owner, id, false)
		if err != nil {
			return err
		}
		if op.Status != "writing" && op.Status != "verifying" {
			return nil
		}
		if _, err = tx.ExecContext(ctx, `UPDATE ai_write_targets SET reads=0,dueAt=? WHERE userID=? AND installationID=? AND operationID=? AND reads=3`, s.drafts.results.jobs.now().Add(time.Second).UnixNano(), owner, op.Plan.Installation, id); err != nil {
			return drafts.ErrStorage
		}
		return nil
	})
	return op, err
}
