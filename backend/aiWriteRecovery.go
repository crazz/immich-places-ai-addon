package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func (a *aiWriteAttempt) recover(ctx context.Context, op writeback.Operation) (bool, error) {
	claimed := false
	err := a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		a.token = uuid.NewString()
		now := a.store.drafts.results.jobs.now()
		result, err := tx.ExecContext(ctx, aiWriteStatement(op.Plan, `UPDATE ai_write_targets SET leaseToken=?,generation=generation+1,leaseUntil=?,reads=reads+1,dueAt=? + CASE reads WHEN 0 THEN 5000000000 ELSE 15000000000 END WHERE userID=? AND installationID=? AND operationID=? AND reads<3 AND dueAt<=? AND (senderActive=0 OR leaseUntil<=?) AND EXISTS(SELECT 1 FROM ai_write_operations o WHERE o.userID=ai_write_targets.userID AND o.installationID=ai_write_targets.installationID AND o.id=ai_write_targets.operationID AND o.status IN ('writing','verifying'))`), a.token, now.Add(45*time.Second).UnixNano(), now.UnixNano(), op.Plan.Owner, op.Plan.Installation, op.ID, now.UnixNano(), now.UnixNano())
		if err != nil {
			return drafts.ErrStorage
		}
		n, err := result.RowsAffected()
		claimed = n == 1
		return err
	})
	return claimed, err
}
