package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/jobs"
)

type aiProductionJobs struct {
	store       *aiJobStore
	selections  *aiSelectionStore
	policies    jobs.ExecutionPolicies
	fingerprint string
}

func (p *aiProductionJobs) submit(ctx context.Context, owner string, req jobs.Admission) (jobs.Job, error) {
	req, requestDigest, err := jobs.NormalizeAdmission(req)
	if err != nil {
		return jobs.Job{}, err
	}
	id := uuid.NewString()
	err = p.store.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := p.store.current(ctx, tx); err != nil {
			return err
		}

		var existingID string
		var existingDigest sql.NullString
		lookupErr := tx.QueryRowContext(ctx, `SELECT j.id,a.requestDigest FROM ai_jobs j LEFT JOIN ai_job_admissions a ON a.userID=j.userID AND a.jobID=j.id WHERE j.userID=? AND j.installationID=? AND j.idempotencyKey=?`, owner, p.store.binding, req.IdempotencyKey).Scan(&existingID, &existingDigest)
		if lookupErr == nil {
			if !existingDigest.Valid || existingDigest.String != requestDigest {
				return jobs.ErrConflict
			}
			id = existingID
			return nil
		}
		if !errors.Is(lookupErr, sql.ErrNoRows) {
			return lookupErr
		}
		manifest, err := p.selections.loadInTx(ctx, tx, owner, req.Configuration.SelectionToken)
		if err != nil {
			return jobs.ErrDenied
		}
		if err := p.validateRerun(ctx, tx, owner, req.Configuration.Rerun, manifest.AssetIDs); err != nil {
			return err
		}
		var model string
		cfg := req.Configuration
		if err := tx.QueryRowContext(ctx, `SELECT v.model FROM ai_provider_profiles p JOIN ai_provider_versions v ON v.userID=p.userID AND v.profileID=p.id AND v.revision=p.activeRevision WHERE p.userID=? AND p.id=? AND p.activeRevision=? AND p.enabled=1`, owner, cfg.ProfileID, cfg.Revision).Scan(&model); err != nil {
			return jobs.ErrDenied
		}
		policy, policyID, ok := p.policies.Resolve(jobs.ExecutionBinding{Owner: owner, Installation: p.store.binding, Profile: cfg.ProfileID, Revision: cfg.Revision, Model: model, EgressFingerprint: p.fingerprint})
		if !ok || policyID != cfg.PolicyID || !policy.Allows(cfg.Limits) || (cfg.Mode != "visual" && !policy.Context) {
			return jobs.ErrDenied
		}
		if err := p.policyAvailable(ctx, tx, owner, policyID); err != nil {
			return err
		}
		if err := p.checkCapability(ctx, tx, owner, cfg); err != nil {
			return err
		}
		var selectionDigest string
		if err := tx.QueryRowContext(ctx, `SELECT digest FROM ai_selection_snapshots WHERE userID=? AND id=?`, owner, cfg.SelectionToken).Scan(&selectionDigest); err != nil {
			return err
		}
		calls := cfg.Limits.MaxCalls
		consentVersion := "visual-v1"
		if cfg.Mode == "context-assisted" || cfg.Mode == "research" {
			consentVersion = "context-v1"
			if cfg.Mode == "research" {
				consentVersion = "research-v1"
			}
			if cfg.Context.AlbumID != "" && cfg.Context.AlbumID != manifest.Scope.AlbumID {
				return jobs.ErrDenied
			}
		}
		if calls == 0 {
			calls = len(manifest.AssetIDs)
		}
		input, digest, err := jobs.Normalize(jobs.Submission{Owner: owner, Installation: p.store.binding, Key: req.IdempotencyKey, Profile: cfg.ProfileID, Revision: cfg.Revision, AssetIDs: manifest.AssetIDs, Languages: cfg.Languages, PrimaryLanguage: cfg.PrimaryLanguage, SelectionDigest: selectionDigest, ConsentVersion: consentVersion, MaxCalls: calls, Mode: cfg.Mode})
		if err != nil {
			return err
		}
		data, _ := json.Marshal(input)
		if err := p.store.insertJob(ctx, tx, id, input, digest, string(data), model); err != nil {
			return err
		}
		if err := p.retainLaunch(ctx, tx, owner, id, manifest); err != nil {
			return err
		}
		requestJSON, _ := json.Marshal(req)
		selectionJSON, _ := json.Marshal(manifest)
		policyJSON, _ := json.Marshal(policy)
		_, err = tx.ExecContext(ctx, `INSERT INTO ai_job_admissions VALUES(?,?,'production-v1',?,?,?,?,?)`, owner, id, requestDigest, string(requestJSON), string(selectionJSON), policyID, string(policyJSON))
		return err
	})
	if err != nil {
		return jobs.Job{}, err
	}
	return p.store.Get(ctx, owner, id)
}
