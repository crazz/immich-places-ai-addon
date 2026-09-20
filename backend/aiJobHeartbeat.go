package main

import (
	"context"
	"database/sql"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func (s *aiJobStore) Authorize(ctx context.Context, lease jobs.Lease) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tx, err := s.db.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return aiJobFailure(ctx, err)
	}
	defer tx.Rollback()
	state, err := s.leaseState(ctx, tx, lease)
	if err != nil {
		return aiJobFailure(ctx, err)
	}
	if state.Canceled || state.Blocked {
		return jobs.ErrDenied
	}
	return aiJobFailure(ctx, tx.Commit())
}

func (s *aiJobStore) Heartbeat(ctx context.Context, lease jobs.Lease, policy jobs.Policy) error {
	if !policy.Valid() {
		return jobs.ErrInvalid
	}
	return s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		state, err := s.leaseState(ctx, tx, lease)
		if err != nil {
			return err
		}
		if state.Canceled || state.Blocked {
			return jobs.ErrDenied
		}
		_, err = tx.ExecContext(ctx, `UPDATE ai_job_items SET leaseExpiresAt=? WHERE userID=? AND jobID=? AND id=? AND leaseToken=?`, s.now().Add(policy.LeaseDuration).UnixNano(), lease.Owner, lease.JobID, lease.ItemID, lease.Token)
		return err
	})
}
