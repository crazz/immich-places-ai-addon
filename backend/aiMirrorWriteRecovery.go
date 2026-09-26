package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
)

func (a *aiMirrorAttempt) recover(ctx context.Context) (bool, error) {
	p, s := a.standard.parent, a.standard.store
	claimed := false
	err := s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		a.token = uuid.NewString()
		now := s.drafts.results.jobs.now()
		result, err := tx.ExecContext(ctx, `UPDATE ai_mirror_write_steps SET leaseToken=?,generation=generation+1,leaseUntil=?,reads=reads+1,dueAt=? + CASE reads WHEN 0 THEN 5000000000 ELSE 15000000000 END WHERE userID=? AND installationID=? AND operationID=? AND assetID=? AND status IN ('writing','verifying') AND reads<3 AND dueAt<=? AND (senderActive=0 OR leaseUntil<=?)`, a.token, now.Add(45*time.Second).UnixNano(), now.UnixNano(), p.Plan.Owner, p.Plan.Installation, p.ID, p.Plan.TargetID, now.UnixNano(), now.UnixNano())
		if err != nil {
			return drafts.ErrStorage
		}
		n, err := result.RowsAffected()
		if err != nil {
			return drafts.ErrStorage
		}
		claimed = n == 1
		if claimed && tx.QueryRowContext(ctx, `SELECT guardToken FROM ai_stack_write_targets WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, p.Plan.Owner, p.Plan.Installation, p.ID, p.Plan.TargetID).Scan(&a.guardToken) != nil {
			return drafts.ErrStorage
		}
		return nil
	})
	return claimed, err
}
