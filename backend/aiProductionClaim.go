package main

import (
	"context"

	"immich-places-backend/internal/ai/jobs"
)

type aiProductionWorkerStore struct {
	*aiJobStore
	production *aiProductionJobs
}

func (s *aiProductionWorkerStore) Claim(ctx context.Context, policy jobs.Policy) (jobs.Lease, bool, error) {
	return s.aiJobStore.claim(ctx, policy, true)
}
