package main

import (
	"context"
	"database/sql"

	"immich-places-backend/internal/ai/jobs"
)

type aiJobLeaseState struct {
	Attempts, Calls, JobCalls, MaxCalls, LeaseCalls int
	Canceled, Blocked                               bool
}

func (s *aiJobStore) leaseState(ctx context.Context, tx *sql.Tx, lease jobs.Lease) (aiJobLeaseState, error) {
	var state aiJobLeaseState
	if err := s.current(ctx, tx); err != nil {
		return state, err
	}
	if lease.Installation != s.binding || lease.Token == "" {
		return state, jobs.ErrLease
	}
	err := tx.QueryRowContext(ctx, `SELECT i.attempts,i.calls,j.calls,j.maxCalls,j.cancelRequested,j.blocked,i.leaseCalls FROM ai_job_items i
 JOIN ai_jobs j ON j.userID=i.userID AND j.id=i.jobID
 WHERE i.userID=? AND i.jobID=? AND i.id=? AND i.assetID=? AND j.installationID=? AND i.state='running' AND i.leaseToken=? AND i.leaseExpiresAt>?`, lease.Owner, lease.JobID, lease.ItemID, lease.Asset, s.binding, lease.Token, s.now().UnixNano()).Scan(&state.Attempts, &state.Calls, &state.JobCalls, &state.MaxCalls, &state.Canceled, &state.Blocked, &state.LeaseCalls)
	if err == sql.ErrNoRows {
		return state, jobs.ErrLease
	}
	return state, err
}

func (s *aiJobStore) Reserve(ctx context.Context, lease jobs.Lease) error {
	return s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		state, err := s.leaseState(ctx, tx, lease)
		if err != nil {
			return err
		}
		if state.Canceled || state.Blocked {
			return jobs.ErrDenied
		}
		if state.Calls >= 3 || state.JobCalls >= state.MaxCalls {
			return jobs.ErrBudget
		}
		if _, err = tx.ExecContext(ctx, `UPDATE ai_job_items SET calls=calls+1,leaseCalls=leaseCalls+1 WHERE userID=? AND jobID=? AND id=? AND leaseToken=?`, lease.Owner, lease.JobID, lease.ItemID, lease.Token); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE ai_jobs SET calls=calls+1 WHERE userID=? AND id=?`, lease.Owner, lease.JobID)
		if err != nil {
			return err
		}
		if state.JobCalls+1 == state.MaxCalls {
			_, err = tx.ExecContext(ctx, `UPDATE ai_job_items SET state='failed',failure='budget',finishedAt=? WHERE userID=? AND jobID=? AND state IN ('queued','retry_wait')`, s.now().UnixNano(), lease.Owner, lease.JobID)
		}
		return err
	})
}
