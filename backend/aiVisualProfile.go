package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/providers"
)

type aiVisualProfile struct {
	model, baseURL, attempt, fingerprint string
	credential                           [32]byte
}

func (a *aiVisualAnalyzer) loadProfile(ctx context.Context, req analysis.Request) (aiVisualProfile, error) {
	if !a.dispatcher.Enabled {
		return aiVisualProfile{}, analysis.ErrDenied
	}
	var profile aiVisualProfile
	var enabled bool
	var lifecycle, protocol, observations, requestedModel, secret string
	err := a.db.db.QueryRowContext(ctx, `SELECT p.enabled,v.model,v.baseURL,COALESCE(v.secretCiphertext,''),c.attemptID,c.policyFingerprint,c.lifecycle,c.protocolVersion,c.observationsJSON,c.requestedModel
 FROM ai_provider_profiles p JOIN ai_provider_versions v ON v.userID=p.userID AND v.profileID=p.id AND v.revision=?
 JOIN ai_provider_capability_checks c ON c.userID=p.userID AND c.profileID=p.id AND c.revision=v.revision
 WHERE p.userID=? AND p.id=?`, req.Revision, req.Owner, req.ProfileID).Scan(&enabled, &profile.model, &profile.baseURL, &secret, &profile.attempt, &profile.fingerprint, &lifecycle, &protocol, &observations, &requestedModel)
	if err != nil || !enabled || len(observations) > 8192 || len(secret) > 64<<10 || profile.model != requestedModel {
		return aiVisualProfile{}, analysis.ErrDenied
	}
	if _, err := providers.MatchEgressDestination(a.dispatcher.Policy, profile.baseURL); err != nil {
		return aiVisualProfile{}, analysis.ErrDenied
	}
	var obs capabilities.Observations
	if json.Unmarshal([]byte(observations), &obs) != nil {
		return aiVisualProfile{}, analysis.ErrDenied
	}
	report := capabilities.Report{Revision: req.Revision, ProtocolVersion: protocol, PolicyFingerprint: profile.fingerprint, Lifecycle: lifecycle, Observations: obs}
	if !capabilities.IsApplicable(report, capabilities.ApplicabilityContext{AIEnabled: true, ProfileEnabled: enabled, ActiveRevision: req.Revision, CurrentProtocolVersion: capabilities.ProtocolVersion, CurrentPolicyFingerprint: policyFingerprint(a.dispatcher.Policy)}) || capabilities.CompatibilitySummary(obs) == "failed" {
		return aiVisualProfile{}, analysis.ErrDenied
	}
	if (req.Format == "strict" && obs.Strict.Status != capabilities.StatusSupported) || (req.Format == "json" && (!req.AllowJSON || obs.JSON.Status != capabilities.StatusSupported)) || (req.Format != "strict" && req.Format != "json") {
		return aiVisualProfile{}, analysis.ErrDenied
	}
	profile.credential = sha256.Sum256([]byte(secret))
	return profile, nil
}
