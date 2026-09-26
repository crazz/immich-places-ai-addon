package main

import (
	"context"
	"database/sql"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func (a *aiWriteAttempt) retainVerification(ctx context.Context, op writeback.Operation, fresh writepreview.Metadata) error {
	if op.Plan.Version != "standard-preview-v2" {
		return nil
	}
	// Commit upstream evidence before catalog publication, whose failure must
	// never restore permission to replay an already verified field.
	return a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := a.fence(ctx, tx, op); err != nil {
			return err
		}
		authority, err := a.store.authority(ctx, tx, writeback.Operation{ID: op.ID, Plan: op.Plan, Status: "verifying"})
		if err != nil || authority != a.authority {
			return drafts.ErrUnavailable
		}
		current, err := a.store.read(ctx, tx, op.Plan.Owner, op.ID, false)
		if err != nil {
			return err
		}
		var completed bool
		if tx.QueryRowContext(ctx, aiWriteStatement(op.Plan, `SELECT completionKnown FROM ai_write_targets WHERE userID=? AND installationID=? AND operationID=?`), op.Plan.Owner, op.Plan.Installation, op.ID).Scan(&completed) != nil {
			return drafts.ErrStorage
		}
		decision := writeback.Readback(current, fresh, completed)
		if err = aiStoreWriteFields(ctx, tx, current, decision.Fields); err != nil {
			return err
		}
		return aiRecordAppendLineage(ctx, tx, current, decision.Fields)
	})
}
