package main

import (
	"context"
	"database/sql"

	"immich-places-backend/internal/ai/jobs"
)

func (p *aiProductionJobs) executionAuthority(ctx context.Context, tx *sql.Tx, lease jobs.Lease) (jobs.Admission, jobs.ExecutionPolicy, error) {
	var raw, digest, model, storedPolicy string
	var revision int
	var enabled bool
	err := tx.QueryRowContext(ctx, `SELECT a.requestJSON,a.requestDigest,j.model,p.activeRevision,p.enabled,a.policyID FROM ai_job_admissions a JOIN ai_jobs j ON j.userID=a.userID AND j.id=a.jobID JOIN ai_provider_profiles p ON p.userID=j.userID AND p.id=json_extract(j.requestJSON,'$.Profile') WHERE a.userID=? AND a.jobID=? AND j.installationID=? AND a.version='production-v1'`, lease.Owner, lease.JobID, lease.Installation).Scan(&raw, &digest, &model, &revision, &enabled, &storedPolicy)
	if err == sql.ErrNoRows {
		return jobs.Admission{}, jobs.ExecutionPolicy{}, jobs.ErrDenied
	}
	if err != nil {
		return jobs.Admission{}, jobs.ExecutionPolicy{}, err
	}
	var req jobs.Admission
	if jobs.DecodeRequest([]byte(raw), &req) != nil {
		return req, jobs.ExecutionPolicy{}, jobs.ErrDenied
	}
	req, actual, err := jobs.NormalizeAdmission(req)
	if err != nil || actual != digest || !enabled || revision != req.Configuration.Revision || storedPolicy != req.Configuration.PolicyID {
		return req, jobs.ExecutionPolicy{}, jobs.ErrDenied
	}
	policy, id, ok := p.policies.Resolve(jobs.ExecutionBinding{Owner: lease.Owner, Installation: lease.Installation, Profile: req.Configuration.ProfileID, Revision: revision, Model: model, EgressFingerprint: p.fingerprint})
	if !ok || id != storedPolicy || !policy.Allows(req.Configuration.Limits) || (req.Configuration.Mode != "visual" && !policy.Context) {
		return req, jobs.ExecutionPolicy{}, jobs.ErrDenied
	}
	if err = p.policyAvailable(ctx, tx, lease.Owner, id); err != nil {
		return req, policy, err
	}
	if err = p.checkCapability(ctx, tx, lease.Owner, req.Configuration); err != nil {
		return req, policy, err
	}
	return req, policy, nil
}
func (s *aiProductionWorkerStore) Authorize(ctx context.Context, lease jobs.Lease) error {
	return s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		state, err := s.leaseState(ctx, tx, lease)
		if err != nil {
			return err
		}
		if state.Canceled || state.Blocked {
			return jobs.ErrDenied
		}
		_, _, err = s.production.executionAuthority(ctx, tx, lease)
		return err
	})
}

func (p *aiProductionJobs) policyAvailable(ctx context.Context, tx *sql.Tx, owner, id string) error {
	var violated bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM ai_execution_policy_violations WHERE userID=? AND policyID=?)`, owner, id).Scan(&violated); err != nil {
		return err
	}
	if violated {
		return jobs.ErrDenied
	}
	return nil
}
