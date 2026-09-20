package main

import (
	"context"
	"database/sql"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func (s *aiJobStore) PurgeBefore(ctx context.Context, owner string, cutoff time.Time, limit int) (int, error) {
	if !jobs.CleanupAllowed(owner, cutoff, s.now(), limit) {
		return 0, jobs.ErrInvalid
	}
	deleted := 0
	err := s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `DELETE FROM ai_jobs WHERE userID=? AND id IN (
   SELECT j.id FROM ai_jobs j WHERE j.userID=? AND j.createdAt<? AND NOT EXISTS (
    SELECT 1 FROM ai_job_items i WHERE i.userID=j.userID AND i.jobID=j.id AND
    (i.state NOT IN ('succeeded','failed','canceled') OR i.finishedAt IS NULL OR i.finishedAt>=?))
   ORDER BY j.createdAt,j.id LIMIT ?)`, owner, owner, cutoff.UnixNano(), cutoff.UnixNano(), limit)
		if err != nil {
			return err
		}
		count, err := result.RowsAffected()
		deleted = int(count)
		return err
	})
	if err != nil {
		return 0, err
	}
	return deleted, nil
}
