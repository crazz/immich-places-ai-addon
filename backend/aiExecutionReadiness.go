package main

import (
	"context"
	"database/sql"
	"errors"

	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/jobs"
)

type aiExecutionReadiness struct {
	Status          string `json:"status"`
	Source          string `json:"source,omitempty"`
	PolicyID        string `json:"policyID,omitempty"`
	MaxInputTokens  int64  `json:"maxInputTokens,omitempty"`
	MaxOutputTokens int64  `json:"maxOutputTokens,omitempty"`
	MaxRequestBytes int    `json:"maxRequestBytes,omitempty"`
	MaxImageBytes   int    `json:"maxImageBytes,omitempty"`
	CostStatus      string `json:"costStatus"`
	Currency        string `json:"currency,omitempty"`
}

func (h *aiProviderHandlers) executionReadiness(ctx context.Context, owner string, profile aiProviderProfile) (*aiExecutionReadiness, error) {
	ready := &aiExecutionReadiness{Status: "policy_required", CostStatus: "unknown"}
	if !h.enabled || !profile.Enabled {
		ready.Status = "unavailable"
		return ready, nil
	}
	var installation string
	err := h.db.db.QueryRowContext(ctx, "SELECT id FROM ai_installation_identity WHERE singleton=1").Scan(&installation)
	if errors.Is(err, sql.ErrNoRows) {
		return ready, nil
	}
	if err != nil {
		return nil, err
	}
	p, digest, found := h.executionPolicies.Find(jobs.ExecutionBinding{Owner: owner, Installation: installation, Profile: profile.ID, Revision: profile.Revision, Model: profile.Model, EgressFingerprint: policyFingerprint(h.policy)})
	if !found {
		return ready, nil
	}
	ready.Source, ready.PolicyID = "operator-attested", digest
	ready.MaxInputTokens, ready.MaxOutputTokens = p.MaxInputTokens, p.MaxOutputTokens
	ready.MaxRequestBytes, ready.MaxImageBytes = p.MaxRequestBytes, p.MaxImageBytes
	if p.Currency != "" {
		ready.CostStatus, ready.Currency = "estimated", p.Currency
	}
	var violated bool
	if err := h.db.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM ai_execution_policy_violations WHERE userID=? AND policyID=?)`, owner, digest).Scan(&violated); err != nil {
		return nil, err
	}
	if violated {
		ready.Status = "policy_violated"
		return ready, nil
	}
	ready.Status = "capability_required"
	r := profile.CapabilityReport
	if r != nil && r.Applicable && r.Observations.Image.Status == capabilities.StatusSupported && (r.Observations.Strict.Status == capabilities.StatusSupported || r.Observations.JSON.Status == capabilities.StatusSupported) && capabilities.CompatibilitySummary(r.Observations) != "failed" {
		ready.Status = "ready"
	}
	return ready, nil
}
