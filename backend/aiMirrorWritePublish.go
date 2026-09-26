package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func (a *aiMirrorAttempt) verify(ctx context.Context) error {
	fresh, err := a.read(ctx, true)
	if err != nil {
		return err
	}
	p, s := a.standard.parent, a.standard.store
	return s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := a.fence(ctx, tx); err != nil {
			return err
		}
		authority, err := a.standard.assetAuthority(ctx, tx, p.Plan.TargetID, true)
		if err != nil || authority != a.standard.authority {
			return drafts.ErrUnavailable
		}
		current, err := s.readStackOperation(ctx, tx, p.Plan.Owner, p.ID, false)
		if err != nil {
			return err
		}
		m := current.Mirror
		decision := writeback.MirrorReadback(p.Plan, fresh, m.Settled, m.Attempts, m.Verified)
		observed, err := json.Marshal(fresh.Mirror)
		if err != nil {
			return drafts.ErrStorage
		}
		_, err = tx.ExecContext(ctx, `UPDATE ai_mirror_write_steps SET status=?,code=?,observed=?,verified=max(verified,?),leaseUntil=0,updatedAt=? WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, decision.Status, decision.Code, string(observed), decision.Verified, s.drafts.results.jobs.now().UnixNano(), p.Plan.Owner, p.Plan.Installation, p.ID, p.Plan.TargetID)
		if err != nil {
			return drafts.ErrStorage
		}
		if decision.Verified {
			result, err := tx.ExecContext(ctx, `UPDATE ai_mirror_records SET lastOperationID=?,value=?,verifiedAt=? WHERE userID=? AND installationID=? AND assetID=? AND recordID=?`, p.ID, string(p.Plan.Mirror.Value), s.drafts.results.jobs.now().UnixNano(), p.Plan.Owner, p.Plan.Installation, p.Plan.TargetID, p.Plan.Mirror.RecordID)
			if err != nil {
				return drafts.ErrStorage
			}
			if n, err := result.RowsAffected(); err != nil || n != 1 {
				return drafts.ErrStorage
			}
		}
		if m.Settled && (decision.Status == "succeeded" || decision.Status == "failed" || decision.Status == "conflict") {
			if err = a.release(ctx, tx); err != nil {
				return err
			}
		}
		return a.event(ctx, tx, decision.Code)
	})
}

func (a *aiMirrorAttempt) stop(ctx context.Context, status, code string) error {
	p, s := a.standard.parent, a.standard.store
	return s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `UPDATE ai_mirror_write_steps SET status=?,code=?,updatedAt=? WHERE userID=? AND installationID=? AND operationID=? AND assetID=? AND status IN ('blocked','queued') AND generation=?`, status, code, s.drafts.results.jobs.now().UnixNano(), p.Plan.Owner, p.Plan.Installation, p.ID, p.Plan.TargetID, p.Mirror.Generation)
		if err != nil {
			return drafts.ErrStorage
		}
		if n, err := result.RowsAffected(); err != nil || n != 1 {
			return writeback.Failure("STALE_WORKER")
		}
		if err = a.release(ctx, tx); err != nil {
			return err
		}
		return a.event(ctx, tx, code)
	})
}

func (a *aiMirrorAttempt) release(ctx context.Context, tx *sql.Tx) error {
	p := a.standard.parent
	_, err := tx.ExecContext(ctx, `DELETE FROM ai_write_target_guards WHERE installationID=? AND assetID=? AND token=(SELECT guardToken FROM ai_stack_write_targets WHERE userID=? AND installationID=? AND operationID=? AND assetID=? AND completionKnown=1 AND senderActive=0) AND EXISTS(SELECT 1 FROM ai_mirror_write_steps WHERE userID=? AND installationID=? AND operationID=? AND assetID=? AND completionKnown=1 AND senderActive=0)`, p.Plan.Installation, p.Plan.TargetID, p.Plan.Owner, p.Plan.Installation, p.ID, p.Plan.TargetID, p.Plan.Owner, p.Plan.Installation, p.ID, p.Plan.TargetID)
	return err
}
