package main

import (
	"context"
	"database/sql"

	"immich-places-backend/internal/ai/jobs"
)

func (s *aiProductionWorkerStore) Reserve(ctx context.Context, lease jobs.Lease) error {
	return s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		state, err := s.leaseState(ctx, tx, lease)
		if err != nil {
			return err
		}
		if state.Canceled || state.Blocked {
			return jobs.ErrDenied
		}
		if state.LeaseCalls != 0 {
			return jobs.ErrBudget
		}
		req, policy, err := s.production.executionAuthority(ctx, tx, lease)
		if err != nil {
			return err
		}
		var consumed int64
		if err = tx.QueryRowContext(ctx, `SELECT COALESCE(sum(inputReserved+outputReserved),0) FROM ai_job_usage WHERE userID=? AND jobID=?`, lease.Owner, lease.JobID).Scan(&consumed); err != nil {
			return err
		}
		allowance := policy.MaxInputTokens + req.Configuration.Limits.OutputTokens
		if allowance > req.Configuration.Limits.MaxTokens-consumed {
			return jobs.ErrBudget
		}

		estimate := policy.EstimatedMicros(req.Configuration.Limits.OutputTokens)
		var previous int64
		if cap := req.Configuration.Limits.MaxEstimatedMicros; cap != nil {
			if err = tx.QueryRowContext(ctx, `SELECT COALESCE(sum(estimatedMicros),0) FROM ai_job_usage WHERE userID=? AND jobID=?`, lease.Owner, lease.JobID).Scan(&previous); err != nil {
				return err
			}
			if estimate == nil || *estimate > *cap-previous {
				return jobs.ErrBudget
			}
		}
		if err = s.reserveInTx(ctx, tx, lease); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO ai_job_usage(userID,jobID,itemID,leaseToken,policyID,inputReserved,outputReserved,estimatedMicros,createdAt) VALUES(?,?,?,?,?,?,?,?,?)`, lease.Owner, lease.JobID, lease.ItemID, lease.Token, req.Configuration.PolicyID, policy.MaxInputTokens, req.Configuration.Limits.OutputTokens, estimate, s.now().UnixNano())

		if err != nil {
			return err
		}
		exhausted := req.Configuration.Limits.MaxTokens-consumed-allowance < allowance
		if cap := req.Configuration.Limits.MaxEstimatedMicros; cap != nil && estimate != nil {
			exhausted = exhausted || *cap-previous-*estimate < *estimate
		}
		if exhausted {
			_, err = tx.ExecContext(ctx, `UPDATE ai_job_items SET state='failed',failure='budget',finishedAt=? WHERE userID=? AND jobID=? AND state IN ('queued','retry_wait')`, s.now().UnixNano(), lease.Owner, lease.JobID)
		}
		return err
	})
}
