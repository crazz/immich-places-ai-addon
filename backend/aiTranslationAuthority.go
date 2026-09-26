package main

import (
	"context"
	"database/sql"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/translations"
)

func (s *aiTranslationStore) authority(ctx context.Context, tx *sql.Tx, owner string, req translations.Request) (jobs.ExecutionPolicy, string, error) {
	var model string
	err := tx.QueryRowContext(ctx, `SELECT v.model FROM ai_provider_profiles p JOIN ai_provider_versions v ON v.userID=p.userID AND v.profileID=p.id AND v.revision=p.activeRevision WHERE p.userID=? AND p.id=? AND p.activeRevision=? AND p.enabled=1`, owner, req.ProfileID, req.ProfileRevision).Scan(&model)
	if err != nil || !s.drafts.results.jobs.enabled {
		return jobs.ExecutionPolicy{}, "", drafts.ErrUnavailable
	}
	policy, id, ok := s.policies.Resolve(jobs.ExecutionBinding{Owner: owner, Installation: s.drafts.results.jobs.binding, Profile: req.ProfileID, Revision: req.ProfileRevision, Model: model, EgressFingerprint: s.fingerprint})
	if !ok {
		return jobs.ExecutionPolicy{}, "", drafts.ErrUnavailable
	}
	var violated bool
	if tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM ai_execution_policy_violations WHERE userID=? AND policyID=?)`, owner, id).Scan(&violated) != nil {
		return jobs.ExecutionPolicy{}, "", drafts.ErrStorage
	}
	if violated {
		return jobs.ExecutionPolicy{}, "", drafts.ErrUnavailable
	}
	return policy, id, nil
}

func (s *aiTranslationStore) authorizeItem(ctx context.Context, tx *sql.Tx, owner string, run aiTranslationRun, tag string) error {
	_, policyID, err := s.authority(ctx, tx, owner, run.Request)
	if err != nil || policyID != run.PolicyID {
		return drafts.ErrUnavailable
	}
	var state string
	if tx.QueryRowContext(ctx, `SELECT state FROM ai_translation_items WHERE userID=? AND installationID=? AND runID=? AND language=? AND cancelRequested=0`, owner, s.drafts.results.jobs.binding, run.ID, tag).Scan(&state) != nil || state != "reserved" {
		return drafts.ErrUnavailable
	}
	return nil
}
