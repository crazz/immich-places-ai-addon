package main

import (
	"context"
	"database/sql"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func (s *aiWriteStore) retryMirror(ctx context.Context, owner, id, asset string, generation int) (writeback.Operation, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	parent, err := s.get(ctx, owner, id, false)
	if err != nil {
		return parent, err
	}
	if parent.Mirror == nil || asset != parent.Plan.TargetID || generation < 1 {
		return parent, writeback.Failure("RETRY_UNAVAILABLE")
	}
	repeated := false
	err = s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var from int
		if tx.QueryRowContext(ctx, `SELECT retryFrom FROM ai_mirror_write_steps WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, owner, parent.Plan.Installation, id, asset).Scan(&from) != nil {
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
	if parent.Mirror.Status != "retryable" || parent.Mirror.Generation != generation || !parent.Mirror.Settled || parent.Mirror.Verified || parent.Mirror.Attempts >= 2 {
		return parent, writeback.Failure("RETRY_UNAVAILABLE")
	}
	a := &aiMirrorAttempt{standard: &aiStackAttempt{store: s, parent: parent, assetID: asset}}
	fresh, readErr := a.read(ctx, false)
	step, stepErr := writeback.TargetStep(parent, asset)
	eligible := readErr == nil && stepErr == nil && fresh.Mirror != nil && fresh.ImageIdentity == parent.Plan.ImageIdentity && writeback.EqualMirrorBaseline(*fresh.Mirror, parent.Plan.Mirror.Before) && writeback.Readback(step, fresh, true).Verified
	err = s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, err := s.readStackOperation(ctx, tx, owner, id, false)
		if err != nil {
			return err
		}
		var from int
		if tx.QueryRowContext(ctx, `SELECT retryFrom FROM ai_mirror_write_steps WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, owner, parent.Plan.Installation, id, asset).Scan(&from) != nil {
			return drafts.ErrStorage
		}
		if from == generation {
			parent = current
			return nil
		}
		m := current.Mirror
		if !eligible || m == nil || !writeback.MirrorStandardReady(current) || m.Status != "retryable" || m.Generation != generation || !m.Settled || m.Verified || m.Attempts >= 2 {
			return writeback.Failure("RETRY_UNAVAILABLE")
		}
		if err = a.standard.dispatchAuthority(ctx, tx); err != nil {
			return err
		}
		var guard string
		if tx.QueryRowContext(ctx, `SELECT t.guardToken FROM ai_stack_write_targets t JOIN ai_write_target_guards g ON g.installationID=t.installationID AND g.assetID=t.assetID AND g.token=t.guardToken WHERE t.userID=? AND t.installationID=? AND t.operationID=? AND t.assetID=?`, owner, parent.Plan.Installation, id, asset).Scan(&guard) != nil {
			return writeback.Failure("TARGET_BUSY")
		}
		if _, err = tx.ExecContext(ctx, `UPDATE ai_mirror_write_steps SET retryFrom=?,generation=generation+1,reads=0,dueAt=0,status='queued',code='EXPLICIT_RETRY' WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, generation, owner, parent.Plan.Installation, id, asset); err != nil {
			return drafts.ErrStorage
		}
		if err = a.event(ctx, tx, "EXPLICIT_RETRY"); err != nil {
			return err
		}
		parent, err = s.readStackOperation(ctx, tx, owner, id, false)
		return err
	})
	return parent, err
}
