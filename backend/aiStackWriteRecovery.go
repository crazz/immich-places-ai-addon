package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
)

func (a *aiStackAttempt) recover(ctx context.Context) (bool, error) {
	claimed := false
	err := a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		a.token = uuid.NewString()
		now := a.store.drafts.results.jobs.now()
		result, err := tx.ExecContext(ctx, `UPDATE ai_stack_write_targets SET leaseToken=?,generation=generation+1,leaseUntil=?,reads=reads+1,dueAt=? + CASE reads WHEN 0 THEN 5000000000 ELSE 15000000000 END WHERE userID=? AND installationID=? AND operationID=? AND assetID=? AND status IN ('writing','verifying') AND reads<3 AND dueAt<=? AND (senderActive=0 OR leaseUntil<=?)`, a.token, now.Add(45*time.Second).UnixNano(), now.UnixNano(), a.parent.Plan.Owner, a.parent.Plan.Installation, a.parent.ID, a.assetID, now.UnixNano(), now.UnixNano())
		if err != nil {
			return drafts.ErrStorage
		}
		n, err := result.RowsAffected()
		if err != nil {
			return drafts.ErrStorage
		}
		claimed = n == 1
		if claimed && tx.QueryRowContext(ctx, `SELECT guardToken FROM ai_stack_write_targets WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, a.parent.Plan.Owner, a.parent.Plan.Installation, a.parent.ID, a.assetID).Scan(&a.guardToken) != nil {
			return drafts.ErrStorage
		}
		return nil
	})
	return claimed, err
}
