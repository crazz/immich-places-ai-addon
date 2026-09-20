package main

import (
	"context"
	"database/sql"

	"immich-places-backend/internal/ai/jobs"
)

func (s *aiProductionWorkerStore) Complete(ctx context.Context, lease jobs.Lease, completion jobs.Completion) (string, error) {
	return s.aiJobStore.complete(ctx, lease, completion, func(ctx context.Context, tx *sql.Tx) error {
		_, _, err := s.production.executionAuthority(ctx, tx, lease)
		return err
	})
}
