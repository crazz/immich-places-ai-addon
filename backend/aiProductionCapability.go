package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/jobs"
)

func (p *aiProductionJobs) checkCapability(ctx context.Context, tx *sql.Tx, owner string, cfg jobs.Configuration) error {
	var report capabilities.Report
	var observations, requestedModel, model string
	err := tx.QueryRowContext(ctx, `SELECT c.lifecycle,c.protocolVersion,c.policyFingerprint,c.observationsJSON,c.requestedModel,v.model FROM ai_provider_capability_checks c JOIN ai_provider_versions v ON v.userID=c.userID AND v.profileID=c.profileID AND v.revision=c.revision WHERE c.userID=? AND c.profileID=? AND c.revision=?`, owner, cfg.ProfileID, cfg.Revision).Scan(&report.Lifecycle, &report.ProtocolVersion, &report.PolicyFingerprint, &observations, &requestedModel, &model)
	if err != nil || len(observations) > 8192 || model != requestedModel || json.Unmarshal([]byte(observations), &report.Observations) != nil {
		return jobs.ErrDenied
	}
	report.Revision = cfg.Revision
	if !capabilities.IsApplicable(report, capabilities.ApplicabilityContext{AIEnabled: true, ProfileEnabled: true, ActiveRevision: cfg.Revision, CurrentProtocolVersion: capabilities.ProtocolVersion, CurrentPolicyFingerprint: p.fingerprint}) || report.Observations.Image.Status != capabilities.StatusSupported || capabilities.CompatibilitySummary(report.Observations) == "failed" {
		return jobs.ErrDenied
	}
	if (cfg.Format == "strict" && report.Observations.Strict.Status != capabilities.StatusSupported) || (cfg.Format == "json" && report.Observations.JSON.Status != capabilities.StatusSupported) {
		return jobs.ErrDenied
	}
	return nil
}
