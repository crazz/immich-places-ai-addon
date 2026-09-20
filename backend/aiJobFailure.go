package main

import (
	"context"
	"database/sql"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func (s *aiJobStore) Fail(ctx context.Context, lease jobs.Lease, failure jobs.Failure, jitter time.Duration) error {
	return s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		state, err := s.leaseState(ctx, tx, lease)
		if err != nil {
			return err
		}
		next := jobs.AfterFailure(state.Attempts, state.Calls, state.JobCalls, state.MaxCalls, failure, jitter)
		if state.Blocked {
			next = jobs.Transition{State: "blocked", Failure: "blocked"}
		}
		if state.Canceled {
			next = jobs.Transition{State: "canceled", Failure: "canceled"}
		}
		now := s.now()
		if next.State == "blocked" {
			if _, err = tx.ExecContext(ctx, `UPDATE ai_jobs SET blocked=1 WHERE userID=? AND id=?`, lease.Owner, lease.JobID); err != nil {
				return err
			}
			if _, err = tx.ExecContext(ctx, `UPDATE ai_job_items SET state='blocked',failure='blocked',finishedAt=? WHERE userID=? AND jobID=? AND state IN ('queued','retry_wait')`, now.UnixNano(), lease.Owner, lease.JobID); err != nil {
				return err
			}
		}
		var finished *int64
		if next.State != "retry_wait" {
			value := now.UnixNano()
			finished = &value
		}
		_, err = tx.ExecContext(ctx, `UPDATE ai_job_items SET state=?,failure=?,nextAttemptAt=?,leaseToken=NULL,leaseExpiresAt=NULL,finishedAt=? WHERE userID=? AND jobID=? AND id=? AND leaseToken=?`, next.State, next.Failure, now.Add(next.Delay).UnixNano(), finished, lease.Owner, lease.JobID, lease.ItemID, lease.Token)
		return err
	})
}
