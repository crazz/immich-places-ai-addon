package main

import (
	"context"

	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/providers"
)

type aiProviderProfile struct {
	providers.Config
	ID               string               `json:"id"`
	Revision         int                  `json:"revision"`
	Enabled          bool                 `json:"enabled"`
	HasSecret        bool                 `json:"hasSecret"`
	CapabilityReport *capabilities.Report `json:"capabilityReport,omitempty"`
}

type aiProviderInputError struct{ message string }

func (e *aiProviderInputError) Error() string { return e.message }

func (d *Database) createAIProvider(ctx context.Context, userID, id string, input providers.Input) (aiProviderProfile, error) {
	input, err := providers.Validate(input)
	if err != nil {
		return aiProviderProfile{}, &aiProviderInputError{err.Error()}
	}
	var ciphertext *string
	if input.Secret != nil && *input.Secret != "" {
		encrypted, err := encryptValue(d.encryptionKey, *input.Secret)
		if err != nil {
			return aiProviderProfile{}, err
		}
		ciphertext = &encrypted
	}
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return aiProviderProfile{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO ai_provider_profiles (userID, id, activeRevision, enabled) VALUES (?, ?, 1, ?)`, userID, id, input.Enabled); err != nil {
		return aiProviderProfile{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO ai_provider_versions (userID, profileID, revision, name, baseURL, model, secretCiphertext) VALUES (?, ?, 1, ?, ?, ?, ?)`, userID, id, input.Name, input.BaseURL, input.Model, ciphertext); err != nil {
		return aiProviderProfile{}, err
	}
	if err := tx.Commit(); err != nil {
		return aiProviderProfile{}, err
	}
	return aiProviderProfile{Config: input.Config, ID: id, Revision: 1, Enabled: input.Enabled, HasSecret: ciphertext != nil}, nil
}

func (d *Database) listAIProviders(ctx context.Context, userID string, apply capabilities.ApplicabilityContext) ([]aiProviderProfile, error) {
	rows, err := d.db.QueryContext(ctx, `SELECT p.id, p.activeRevision, p.enabled, v.name, v.baseURL, v.model, v.secretCiphertext IS NOT NULL
		FROM ai_provider_profiles p JOIN ai_provider_versions v ON v.userID = p.userID AND v.profileID = p.id AND v.revision = p.activeRevision
		WHERE p.userID = ? ORDER BY v.name, p.id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	profiles := make([]aiProviderProfile, 0)
	for rows.Next() {
		var profile aiProviderProfile
		if err := rows.Scan(&profile.ID, &profile.Revision, &profile.Enabled, &profile.Name, &profile.BaseURL, &profile.Model, &profile.HasSecret); err != nil {
			return nil, err
		}
		report, err := d.loadAIProviderCapability(ctx, userID, profile.ID, profile.Revision, capabilities.ApplicabilityContext{
			AIEnabled:                apply.AIEnabled,
			ProfileEnabled:           profile.Enabled,
			ActiveRevision:           profile.Revision,
			CurrentProtocolVersion:   capabilities.ProtocolVersion,
			CurrentPolicyFingerprint: apply.CurrentPolicyFingerprint,
		})
		if err != nil {
			return nil, err
		}
		profile.CapabilityReport = report
		profiles = append(profiles, profile)
	}
	return profiles, rows.Err()
}
