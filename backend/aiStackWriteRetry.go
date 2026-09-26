package main

import (
	"context"
	"database/sql"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func (s *aiWriteStore) retryStackTarget(ctx context.Context, owner, id, asset string, generation int) (writeback.Operation, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	parent, err := s.get(ctx, owner, id, false)
	if err != nil {
		return parent, err
	}
	step, err := writeback.TargetStep(parent, asset)
	if err != nil || generation < 1 {
		return parent, writeback.Failure("RETRY_UNAVAILABLE")
	}
	a := &aiStackAttempt{store: s, parent: parent, assetID: asset}
	repeated := false
	err = s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var from int
		if tx.QueryRowContext(ctx, `SELECT retryFrom FROM ai_stack_write_targets WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, owner, parent.Plan.Installation, id, asset).Scan(&from) != nil {
			return drafts.ErrUnavailable
		}
		repeated = from == generation
		if repeated {
			parent, err = s.readStackOperation(ctx, tx, owner, id, false)
		}
		return err
	})
	if err != nil || repeated {
		return parent, err
	}
	step.Status = "queued"
	fresh, readErr := a.Read(ctx, step)
	eligible := readErr == nil && writeback.CompareBefore(step.Plan, fresh) == ""
	err = s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, err := a.current(ctx, tx)
		if err != nil {
			return err
		}
		var from int
		var known, active bool
		var token string
		if tx.QueryRowContext(ctx, `SELECT retryFrom,completionKnown,senderActive,guardToken FROM ai_stack_write_targets WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, owner, parent.Plan.Installation, id, asset).Scan(&from, &known, &active, &token) != nil {
			return drafts.ErrStorage
		}
		if from == generation {
			parent, err = s.readStackOperation(ctx, tx, owner, id, false)
			return err
		}
		if !eligible || current.Status != "retryable" || current.Generation != generation || !known || active || current.Attempts >= 2 {
			return writeback.Failure("RETRY_UNAVAILABLE")
		}
		for _, field := range current.Fields {
			if field.WasVerified || field.Status == "verified" {
				return writeback.Failure("RETRY_UNAVAILABLE")
			}
		}
		if err := a.dispatchAuthority(ctx, tx); err != nil {
			return err
		}
		var guard string
		if tx.QueryRowContext(ctx, `SELECT token FROM ai_write_target_guards WHERE installationID=? AND assetID=?`, parent.Plan.Installation, asset).Scan(&guard) != nil || guard != token {
			return writeback.Failure("TARGET_BUSY")
		}
		if _, err := tx.ExecContext(ctx, `UPDATE ai_stack_write_targets SET retryFrom=?,generation=generation+1,reads=0,dueAt=0,status='queued',code='EXPLICIT_RETRY' WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, generation, owner, parent.Plan.Installation, id, asset); err != nil {
			return drafts.ErrStorage
		}
		if err := a.event(ctx, tx, "EXPLICIT_RETRY", current.Attempts); err != nil {
			return err
		}
		parent, err = s.readStackOperation(ctx, tx, owner, id, false)
		return err
	})
	return parent, err
}
