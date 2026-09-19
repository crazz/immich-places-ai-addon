package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

// newAIProviderCapabilityTransport bounds capability probe responses more tightly than the
// general provider transport default.
func newAIProviderCapabilityTransport(opts providerhttp.Options) *providerhttp.Client {
	opts.MaxResponseBytes = providerhttp.MaxCapabilityResponseBytes
	return providerhttp.New(opts)
}

func newAIProviderDispatcher(db *Database, enabled bool, policy providers.EgressPolicy, transport providers.Transport) *providers.Dispatcher {
	return &providers.Dispatcher{
		Enabled:   enabled,
		Policy:    policy,
		Store:     &aiProviderDispatchStore{db: db},
		Transport: transport,
	}
}

type aiProviderDispatchStore struct {
	db *Database
}

func (s *aiProviderDispatchStore) LoadDispatchVersion(ctx context.Context, ownerID, profileID string, revision int) (providers.ProfileVersion, error) {
	var version providers.ProfileVersion
	var hasSecret bool
	err := s.db.db.QueryRowContext(ctx, `
		SELECT p.userID, p.id, v.revision, v.name, v.baseURL, v.model, p.enabled, v.secretCiphertext IS NOT NULL
		FROM ai_provider_profiles p
		JOIN ai_provider_versions v ON v.userID = p.userID AND v.profileID = p.id AND v.revision = ?
		WHERE p.userID = ? AND p.id = ?`, revision, ownerID, profileID).
		Scan(&version.OwnerID, &version.ProfileID, &version.Revision, &version.Name, &version.BaseURL, &version.Model, &version.Enabled, &hasSecret)
	if errors.Is(err, sql.ErrNoRows) {
		return providers.ProfileVersion{}, providers.ErrUnavailable
	}
	if err != nil {
		return providers.ProfileVersion{}, err
	}
	version.HasSecret = hasSecret
	return version, nil
}

func (s *aiProviderDispatchStore) AdmitDispatch(ctx context.Context, ownerID, profileID string, revision int) (providers.DispatchAdmission, error) {
	var enabled bool
	var ciphertext sql.NullString
	err := s.db.db.QueryRowContext(ctx, `
		SELECT p.enabled, v.secretCiphertext
		FROM ai_provider_profiles p
		JOIN ai_provider_versions v ON v.userID = p.userID AND v.profileID = p.id AND v.revision = ?
		WHERE p.userID = ? AND p.id = ?`, revision, ownerID, profileID).
		Scan(&enabled, &ciphertext)
	if errors.Is(err, sql.ErrNoRows) {
		return providers.DispatchAdmission{}, providers.ErrUnavailable
	}
	if err != nil {
		return providers.DispatchAdmission{}, err
	}
	if !enabled {
		return providers.DispatchAdmission{}, providers.ErrDisabled
	}
	if !ciphertext.Valid {
		return providers.DispatchAdmission{}, nil
	}
	if !strings.HasPrefix(ciphertext.String, encryptedPrefix) {
		return providers.DispatchAdmission{}, providers.ErrCredential
	}
	plain, err := decryptValue(s.db.encryptionKey, ciphertext.String)
	if err != nil {
		return providers.DispatchAdmission{}, providers.ErrCredential
	}
	if strings.ContainsAny(plain, "\r\n") {
		return providers.DispatchAdmission{}, providers.ErrCredential
	}
	return providers.DispatchAdmission{Authorization: "Bearer " + plain}, nil
}
