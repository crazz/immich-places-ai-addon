package main

import (
	"context"
	"database/sql"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/jobs"
)

func (p *aiProductionJobs) recordUsage(ctx context.Context, lease jobs.Lease, usage *analysis.Usage) error {
	if usage == nil {
		return nil
	}
	violated := false
	err := p.store.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := p.store.leaseState(ctx, tx, lease); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `UPDATE ai_job_usage SET inputReported=?,outputReported=?,totalReported=? WHERE userID=? AND jobID=? AND itemID=? AND leaseToken=? AND inputReported IS NULL AND outputReported IS NULL AND totalReported IS NULL`, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, lease.Owner, lease.JobID, lease.ItemID, lease.Token)
		if err != nil {
			return err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if n != 1 {
			return jobs.ErrLease
		}

		var input, output int64
		var policyID, policyVersion string
		if err = tx.QueryRowContext(ctx, `SELECT u.inputReserved,u.outputReserved,u.policyID,json_extract(a.policyJSON,'$.version') FROM ai_job_usage u JOIN ai_job_admissions a ON a.userID=u.userID AND a.jobID=u.jobID WHERE u.userID=? AND u.jobID=? AND u.itemID=? AND u.leaseToken=?`, lease.Owner, lease.JobID, lease.ItemID, lease.Token).Scan(&input, &output, &policyID, &policyVersion); err != nil {
			return err
		}
		if policyVersion == jobs.DefaultExecutionVersion {
			return nil
		}
		violated = (usage.PromptTokens != nil && int64(*usage.PromptTokens) > input) || (usage.CompletionTokens != nil && int64(*usage.CompletionTokens) > output) || (usage.TotalTokens != nil && int64(*usage.TotalTokens) > input+output)
		if violated {
			if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO ai_execution_policy_violations VALUES(?,?,?)`, lease.Owner, policyID, p.store.now().UnixNano()); err != nil {
				return err
			}
			if _, err = tx.ExecContext(ctx, `UPDATE ai_jobs SET blocked=1 WHERE userID=? AND id=?`, lease.Owner, lease.JobID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if violated {
		return jobs.ErrDenied
	}
	return nil
}
