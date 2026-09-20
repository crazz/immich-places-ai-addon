package main

import (
	"context"
	"database/sql"

	"immich-places-backend/internal/ai/jobs"
)

func (s *aiJobStore) Cancel(ctx context.Context, owner, id string) error {
	return s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := s.currentInstallation(ctx, tx); err != nil {
			return err
		}
		changed, err := tx.ExecContext(ctx, `UPDATE ai_jobs SET cancelRequested=1 WHERE userID=? AND id=? AND installationID=?`, owner, id, s.binding)
		if err != nil {
			return err
		}
		count, err := changed.RowsAffected()
		if err != nil {
			return err
		}
		if count != 1 {
			return jobs.ErrDenied
		}
		_, err = tx.ExecContext(ctx, `UPDATE ai_job_items SET state='canceled',failure='canceled',finishedAt=?,leaseToken=NULL,leaseExpiresAt=NULL WHERE userID=? AND jobID=? AND state IN ('queued','retry_wait','blocked')`, s.now().UnixNano(), owner, id)
		return err
	})
}
