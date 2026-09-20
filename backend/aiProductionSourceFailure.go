package main

import (
	"context"
	"errors"
	"strings"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/selection"
)

func (p *aiProductionJobs) sourceFailure(ctx context.Context, lease jobs.Lease, err error) error {
	if !errors.Is(err, analysis.ErrDenied) && !errors.Is(err, errAIImageDenied) {
		return nil
	}
	s := &aiProductionWorkerStore{aiJobStore: p.store, production: p}
	if err := s.Authorize(ctx, lease); err != nil {
		return productionExecutionFailure(err)
	}
	user, err := p.store.db.getUserByID(ctx, lease.Owner)
	if err != nil {
		return jobs.ErrStorage
	}
	if user == nil || user.ImmichAPIKey == nil || strings.TrimSpace(*user.ImmichAPIKey) == "" {
		return jobs.Failure{Code: jobs.Blocked}
	}
	row, err := p.store.db.getAssetByID(ctx, lease.Owner, lease.Asset)
	if err != nil {
		return jobs.ErrStorage
	}
	if row == nil {
		return jobs.Failure{Code: jobs.Permanent}
	}
	candidate := selection.Candidate{Available: true, Type: row.Type, Hidden: row.IsHidden, StackChild: row.StackPrimaryAssetID != nil && *row.StackPrimaryAssetID != "" && *row.StackPrimaryAssetID != lease.Asset, InScope: true}
	if selection.ExclusionReason(candidate) != "" {
		return jobs.Failure{Code: jobs.Permanent}
	}
	return nil
}
