package main

import (
	"context"
	"database/sql"

	"immich-places-backend/internal/ai/jobs"
)

func (p *aiProductionJobs) validateRerun(ctx context.Context, tx *sql.Tx, owner string, choice *jobs.RerunChoice, assets []string) error {
	if choice == nil {
		return nil
	}
	for _, asset := range assets {
		var state string
		err := tx.QueryRowContext(ctx, `SELECT i.state FROM ai_job_items i JOIN ai_jobs j ON j.userID=i.userID AND j.id=i.jobID JOIN ai_job_admissions a ON a.userID=j.userID AND a.jobID=j.id WHERE i.userID=? AND i.jobID=? AND i.assetID=? AND j.installationID=?`, owner, choice.ParentJobID, asset, p.store.binding).Scan(&state)
		if err == sql.ErrNoRows {
			return jobs.ErrDenied
		}
		if err != nil {
			return err
		}
		if choice.Kind == "retry-failed" && state != "failed" {
			return jobs.ErrDenied
		}
	}
	return nil
}
