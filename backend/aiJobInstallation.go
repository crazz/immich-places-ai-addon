package main

import (
	"context"
	"database/sql"
	"time"
)

func invalidateAIJobInstallation(ctx context.Context, tx *sql.Tx, previous string, now time.Time) error {
	if previous == "" {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `UPDATE ai_jobs SET cancelRequested=1 WHERE installationID=? AND EXISTS (SELECT 1 FROM ai_job_items i WHERE i.userID=ai_jobs.userID AND i.jobID=ai_jobs.id AND i.state IN ('queued','running','retry_wait','blocked'))`, previous); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `UPDATE ai_job_items SET state='canceled',failure='installation',leaseToken=NULL,leaseExpiresAt=NULL,finishedAt=? WHERE state IN ('queued','running','retry_wait','blocked') AND EXISTS (SELECT 1 FROM ai_jobs j WHERE j.userID=ai_job_items.userID AND j.id=ai_job_items.jobID AND j.installationID=?)`, now.UnixNano(), previous)
	return err
}
