package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/providers"
)

type capabilityAdmission struct {
	AttemptID         string
	ProfileID         string
	Revision          int
	Model             string
	PolicyFingerprint string
	StartedAt         time.Time
	DeadlineAt        time.Time
}

func (d *Database) admitAIProviderCapability(ctx context.Context, userID, profileID string, expectedRevision int, policyFingerprint string, now, deadline time.Time) (capabilityAdmission, error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return capabilityAdmission{}, err
	}
	defer tx.Rollback()

	var revision int
	var enabled bool
	var model string
	err = tx.QueryRowContext(ctx, `
		SELECT p.activeRevision, p.enabled, v.model
		FROM ai_provider_profiles p
		JOIN ai_provider_versions v ON v.userID = p.userID AND v.profileID = p.id AND v.revision = p.activeRevision
		WHERE p.userID = ? AND p.id = ?`, userID, profileID).Scan(&revision, &enabled, &model)
	if errors.Is(err, sql.ErrNoRows) {
		return capabilityAdmission{}, errAIProviderNotFound
	}
	if err != nil {
		return capabilityAdmission{}, err
	}
	if !enabled {
		return capabilityAdmission{}, capabilities.ErrDisabled
	}
	if revision != expectedRevision {
		return capabilityAdmission{}, errAIProviderConflict
	}

	var existingLifecycle sql.NullString
	var existingDeadline sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT lifecycle, deadlineAt FROM ai_provider_capability_checks
		WHERE userID = ? AND profileID = ? AND revision = ?`, userID, profileID, revision).
		Scan(&existingLifecycle, &existingDeadline)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return capabilityAdmission{}, err
	}
	if err == nil && existingLifecycle.String == "running" {
		if deadlineAt, parseErr := time.Parse(time.RFC3339Nano, existingDeadline.String); parseErr == nil && deadlineAt.After(now) {
			return capabilityAdmission{}, capabilities.ErrBusy
		}
	}

	attemptID := uuid.NewString()
	obs, err := json.Marshal(capabilities.EmptyObservations())
	if err != nil {
		return capabilityAdmission{}, err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO ai_provider_capability_checks (
			userID, profileID, revision, attemptID, protocolVersion, policyFingerprint,
			lifecycle, startedAt, deadlineAt, requestedModel, observationsJSON, compatibility
		) VALUES (?, ?, ?, ?, ?, ?, 'running', ?, ?, ?, ?, 'incomplete')
		ON CONFLICT(userID, profileID, revision) DO UPDATE SET
			attemptID = excluded.attemptID,
			protocolVersion = excluded.protocolVersion,
			policyFingerprint = excluded.policyFingerprint,
			lifecycle = 'running',
			startedAt = excluded.startedAt,
			deadlineAt = excluded.deadlineAt,
			completedAt = NULL,
			requestedModel = excluded.requestedModel,
			reportedModel = NULL,
			observationsJSON = excluded.observationsJSON,
			compatibility = excluded.compatibility`,
		userID, profileID, revision, attemptID, capabilities.ProtocolVersion, policyFingerprint,
		now.UTC().Format(time.RFC3339Nano), deadline.UTC().Format(time.RFC3339Nano), model, string(obs))
	if err != nil {
		return capabilityAdmission{}, err
	}
	if err := tx.Commit(); err != nil {
		return capabilityAdmission{}, err
	}
	return capabilityAdmission{
		AttemptID:         attemptID,
		ProfileID:         profileID,
		Revision:          revision,
		Model:             model,
		PolicyFingerprint: policyFingerprint,
		StartedAt:         now.UTC(),
		DeadlineAt:        deadline.UTC(),
	}, nil
}

func (d *Database) loadAIProviderCapabilityDestination(ctx context.Context, userID, profileID string) (string, error) {
	var baseURL string
	err := d.db.QueryRowContext(ctx, `
		SELECT v.baseURL
		FROM ai_provider_profiles p
		JOIN ai_provider_versions v ON v.userID = p.userID AND v.profileID = p.id AND v.revision = p.activeRevision
		WHERE p.userID = ? AND p.id = ?`, userID, profileID).Scan(&baseURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errAIProviderNotFound
	}
	if err != nil {
		return "", err
	}
	return baseURL, nil
}

func (d *Database) completeAIProviderCapability(ctx context.Context, userID string, report capabilities.Report) error {
	obs, err := json.Marshal(report.Observations)
	if err != nil {
		return err
	}
	var completedAt any
	if report.CompletedAt != nil {
		completedAt = report.CompletedAt.UTC().Format(time.RFC3339Nano)
	}
	result, err := d.db.ExecContext(ctx, `
		UPDATE ai_provider_capability_checks
		SET lifecycle = ?, completedAt = ?, reportedModel = ?, observationsJSON = ?, compatibility = ?
		WHERE userID = ? AND profileID = ? AND revision = ? AND attemptID = ?`,
		report.Lifecycle, completedAt, report.ReportedModel, string(obs), report.Compatibility,
		userID, report.ProfileID, report.Revision, report.AttemptID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errAIProviderConflict
	}
	return nil
}

func (d *Database) loadAIProviderCapability(ctx context.Context, userID, profileID string, revision int, apply capabilities.ApplicabilityContext) (*capabilities.Report, error) {
	var report capabilities.Report
	var startedAt, deadlineAt string
	var completedAt sql.NullString
	var reportedModel sql.NullString
	var observations string
	err := d.db.QueryRowContext(ctx, `
		SELECT attemptID, profileID, revision, protocolVersion, policyFingerprint, lifecycle,
			startedAt, deadlineAt, completedAt, requestedModel, reportedModel, observationsJSON, compatibility
		FROM ai_provider_capability_checks
		WHERE userID = ? AND profileID = ? AND revision = ?`, userID, profileID, revision).
		Scan(&report.AttemptID, &report.ProfileID, &report.Revision, &report.ProtocolVersion, &report.PolicyFingerprint, &report.Lifecycle,
			&startedAt, &deadlineAt, &completedAt, &report.RequestedModel, &reportedModel, &observations, &report.Compatibility)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	report.StartedAt, err = time.Parse(time.RFC3339Nano, startedAt)
	if err != nil {
		return nil, err
	}
	report.DeadlineAt, err = time.Parse(time.RFC3339Nano, deadlineAt)
	if err != nil {
		return nil, err
	}
	if completedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, completedAt.String)
		if err != nil {
			return nil, err
		}
		report.CompletedAt = &parsed
	}
	if reportedModel.Valid {
		report.ReportedModel = &reportedModel.String
	}
	if err := json.Unmarshal([]byte(observations), &report.Observations); err != nil {
		return nil, err
	}
	if report.Lifecycle == "running" && !report.DeadlineAt.After(time.Now().UTC()) {
		completedAt := time.Now().UTC()
		_, err := d.db.ExecContext(ctx, `
			UPDATE ai_provider_capability_checks
			SET lifecycle = 'interrupted', completedAt = ?
			WHERE userID = ? AND profileID = ? AND revision = ? AND attemptID = ? AND lifecycle = 'running'`,
			completedAt.Format(time.RFC3339Nano), userID, profileID, revision, report.AttemptID)
		if err != nil {
			return nil, err
		}
		report.Lifecycle = "interrupted"
		report.CompletedAt = &completedAt
	}
	apply.ActiveRevision = revision
	report.Applicable = capabilities.IsApplicable(report, apply)
	return &report, nil
}

func (d *Database) interruptRunningAIProviderCapabilities(ctx context.Context) error {
	completedAt := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := d.db.ExecContext(ctx, `
		UPDATE ai_provider_capability_checks
		SET lifecycle = 'interrupted', completedAt = ?
		WHERE lifecycle = 'running'`, completedAt)
	return err
}

func (d *Database) admitAIProviderCapabilityProbe(ctx context.Context, userID, profileID string, expectedRevision int) error {
	var revision int
	var enabled bool
	err := d.db.QueryRowContext(ctx, `
		SELECT activeRevision, enabled FROM ai_provider_profiles
		WHERE userID = ? AND id = ?`, userID, profileID).Scan(&revision, &enabled)
	if errors.Is(err, sql.ErrNoRows) {
		return capabilities.ErrUnavailable
	}
	if err != nil {
		return err
	}
	if !enabled {
		return capabilities.ErrDisabled
	}
	if revision != expectedRevision {
		return capabilities.ErrUnavailable
	}
	return nil
}

func policyFingerprint(policy providers.EgressPolicy) string {
	return providers.FingerprintPolicy(policy)
}
