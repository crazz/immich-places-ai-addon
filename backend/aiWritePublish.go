package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"slices"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func (a *aiWriteAttempt) Verify(ctx context.Context, op writeback.Operation) error {
	if a.store.sync == nil {
		return drafts.ErrUnavailable
	}
	drain, cancel := context.WithTimeout(ctx, 5*time.Second)
	err := a.store.sync.pauseUserSync(drain, op.Plan.Owner)
	cancel()
	if err != nil {
		a.pending(ctx, op, "LOCAL_REFRESH_PENDING")
		return err
	}
	defer a.store.sync.releaseUserSyncLock(op.Plan.Owner)
	current, err := a.store.get(ctx, op.Plan.Owner, op.ID, false)
	if err != nil {
		return err
	}
	fresh, err := a.Read(ctx, current)
	if err != nil {
		a.pending(ctx, op, "READBACK_UNAVAILABLE")
		return err
	}
	if err = a.publish(ctx, op, fresh); err != nil {
		a.pending(ctx, op, "LOCAL_REFRESH_PENDING")
	}
	return err
}

func (a *aiWriteAttempt) pending(ctx context.Context, op writeback.Operation, code string) {
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	// This records only an unresolved state under the original fence; it cannot
	// recreate deleted resources or publish coordinates after lost authority.
	_ = a.store.drafts.write(cleanup, func(ctx context.Context, tx *sql.Tx) error {
		if err := a.fence(ctx, tx, op); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, aiWriteStatement(op.Plan, `UPDATE ai_write_operations SET status='verifying',code=? WHERE userID=? AND installationID=? AND id=? AND status IN ('writing','verifying')`), code, op.Plan.Owner, op.Plan.Installation, op.ID); err != nil {
			return err
		}
		current, err := a.store.read(ctx, tx, op.Plan.Owner, op.ID, false)
		if err != nil {
			return err
		}
		return a.store.event(ctx, tx, current, code)
	})
}

func (a *aiWriteAttempt) publish(ctx context.Context, op writeback.Operation, fresh writepreview.Metadata) error {
	if err := a.retainVerification(ctx, op, fresh); err != nil {
		return err
	}
	return a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := a.fence(ctx, tx, op); err != nil {
			return err
		}
		currentAuthority, err := a.store.authority(ctx, tx, writeback.Operation{ID: op.ID, Plan: op.Plan, Status: "verifying"})
		if err != nil || currentAuthority != a.authority {
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
		observed, err := json.Marshal(fresh.GPS)
		if err != nil {
			return drafts.ErrStorage
		}
		refreshed := false
		gpsVerified := decision.GPSVerified || (op.Plan.Version == "gps-preview-v1" && decision.Verified)
		if gpsVerified && slices.Contains(op.Plan.Fields, "gps") {
			result, err := tx.ExecContext(ctx, `UPDATE assets SET latitude=?,longitude=? WHERE userID=? AND immichID=?`, *fresh.GPS.Latitude, *fresh.GPS.Longitude, op.Plan.Owner, op.Plan.TargetID)
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
		if _, err = tx.ExecContext(ctx, aiWriteStatement(op.Plan, `UPDATE ai_write_targets SET observed=?,verified=?,refreshed=?,leaseUntil=0,senderActive=0 WHERE userID=? AND installationID=? AND operationID=?`), string(observed), decision.Verified, refreshed, op.Plan.Owner, op.Plan.Installation, op.ID); err != nil {
			return drafts.ErrStorage
		}
		if _, err = tx.ExecContext(ctx, aiWriteStatement(op.Plan, `UPDATE ai_write_operations SET status=?,code=? WHERE userID=? AND installationID=? AND id=?`), decision.Status, decision.Code, op.Plan.Owner, op.Plan.Installation, op.ID); err != nil {
			return drafts.ErrStorage
		}
		if completed && (decision.Status == "succeeded" || decision.Status == "partial" || decision.Status == "conflict" || decision.Status == "failed") {
			if _, err = tx.ExecContext(ctx, `DELETE FROM ai_write_target_guards WHERE token=?`, op.ID); err != nil {
				return drafts.ErrStorage
			}
		}
		return a.store.event(ctx, tx, current, decision.Code)
	})
}
