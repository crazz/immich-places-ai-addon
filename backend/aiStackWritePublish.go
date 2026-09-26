package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func (a *aiStackAttempt) Verify(ctx context.Context, op writeback.Operation) error {
	if a.store.sync == nil {
		return drafts.ErrUnavailable
	}
	drain, cancel := context.WithTimeout(ctx, 5*time.Second)
	err := a.store.sync.pauseUserSync(drain, op.Plan.Owner)
	cancel()
	if err != nil {
		a.pending(ctx, "LOCAL_REFRESH_PENDING")
		return err
	}
	defer a.store.sync.releaseUserSyncLock(op.Plan.Owner)
	var current writeback.Operation
	err = a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		current, err = a.current(ctx, tx)
		return err
	})
	if err != nil {
		return err
	}
	fresh, err := a.Read(ctx, current)
	if err != nil {
		a.pending(ctx, "READBACK_UNAVAILABLE")
		return err
	}
	if err = a.retainVerification(ctx, fresh); err != nil {
		a.pending(ctx, "LOCAL_REFRESH_PENDING")
		return err
	}
	if err = a.publish(ctx, fresh); err != nil {
		a.pending(ctx, "LOCAL_REFRESH_PENDING")
	}
	return err
}

func (a *aiStackAttempt) pending(ctx context.Context, code string) {
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	_ = a.store.drafts.write(cleanup, func(ctx context.Context, tx *sql.Tx) error {
		if err := a.fence(ctx, tx); err != nil {
			return err
		}
		current, err := a.current(ctx, tx)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE ai_stack_write_targets SET status='verifying',code=? WHERE userID=? AND installationID=? AND operationID=? AND assetID=? AND status IN ('writing','verifying')`, code, current.Plan.Owner, current.Plan.Installation, current.ID, a.assetID); err != nil {
			return err
		}
		return a.event(ctx, tx, code, current.Attempts)
	})
}

func (a *aiStackAttempt) publicationState(ctx context.Context, tx *sql.Tx, fresh writepreview.Metadata) (writeback.Operation, writeback.Decision, bool, error) {
	var empty writeback.Operation
	var decision writeback.Decision
	if err := a.fence(ctx, tx); err != nil {
		return empty, decision, false, err
	}
	authority, err := a.assetAuthority(ctx, tx, a.assetID, true)
	if err != nil || authority != a.authority {
		return empty, decision, false, drafts.ErrUnavailable
	}
	current, err := a.current(ctx, tx)
	if err != nil {
		return empty, decision, false, err
	}
	var completed bool
	if tx.QueryRowContext(ctx, `SELECT completionKnown FROM ai_stack_write_targets WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, current.Plan.Owner, current.Plan.Installation, current.ID, a.assetID).Scan(&completed) != nil {
		return empty, decision, false, drafts.ErrStorage
	}
	return current, writeback.Readback(current, fresh, completed), completed, nil
}

func (a *aiStackAttempt) retainVerification(ctx context.Context, fresh writepreview.Metadata) error {
	return a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, decision, _, err := a.publicationState(ctx, tx, fresh)
		if err != nil {
			return err
		}
		if err := aiStoreStackFields(ctx, tx, current, decision.Fields); err != nil {
			return err
		}
		return aiRecordAppendLineage(ctx, tx, current, decision.Fields)
	})
}

func (a *aiStackAttempt) publish(ctx context.Context, fresh writepreview.Metadata) error {
	return a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, decision, completed, err := a.publicationState(ctx, tx, fresh)
		if err != nil {
			return err
		}
		observed, err := json.Marshal(fresh.GPS)
		if err != nil {
			return drafts.ErrStorage
		}
		refreshed := false
		if decision.GPSVerified {
			result, err := tx.ExecContext(ctx, `UPDATE assets SET latitude=?,longitude=? WHERE userID=? AND immichID=?`, *fresh.GPS.Latitude, *fresh.GPS.Longitude, current.Plan.Owner, a.assetID)
			if err != nil {
				return drafts.ErrStorage
			}
			n, err := result.RowsAffected()
			if err != nil {
				return drafts.ErrStorage
			}
			refreshed = n == 1
			if !refreshed {
				decision.Status, decision.Code = "verifying", "LOCAL_REFRESH_PENDING"
			}
		}
		if _, err = tx.ExecContext(ctx, `UPDATE ai_stack_write_targets SET status=?,code=?,observed=?,verified=?,refreshed=?,leaseUntil=0,senderActive=0 WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, decision.Status, decision.Code, string(observed), decision.Verified, refreshed, current.Plan.Owner, current.Plan.Installation, current.ID, a.assetID); err != nil {
			return drafts.ErrStorage
		}
		if completed && (decision.Status == "succeeded" || decision.Status == "partial" || decision.Status == "conflict" || decision.Status == "failed") && !(a.parent.Plan.Mirror != nil && a.assetID == a.parent.Plan.TargetID) {
			if _, err = tx.ExecContext(ctx, `DELETE FROM ai_write_target_guards WHERE installationID=? AND assetID=? AND token=?`, current.Plan.Installation, a.assetID, a.guardToken); err != nil {
				return drafts.ErrStorage
			}
		}
		return a.event(ctx, tx, decision.Code, current.Attempts)
	})
}
