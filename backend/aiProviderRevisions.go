package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"immich-places-backend/internal/ai/providers"
)

var (
	errAIProviderConflict = errors.New("provider was changed; reload before editing")
	errAIProviderNotFound = errors.New("provider not found")
)

func (d *Database) updateAIProvider(ctx context.Context, userID, id string, expectedRevision int, input providers.Input) (aiProviderProfile, error) {
	input, err := providers.Validate(input)
	if err != nil {
		return aiProviderProfile{}, &aiProviderInputError{err.Error()}
	}
	if expectedRevision < 1 {
		return aiProviderProfile{}, &aiProviderInputError{"expectedRevision must be a positive integer"}
	}
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return aiProviderProfile{}, err
	}
	defer tx.Rollback()
	var revision int
	err = tx.QueryRowContext(ctx, `UPDATE ai_provider_profiles SET activeRevision=activeRevision+1, enabled=? WHERE userID=? AND id=? AND activeRevision=? RETURNING activeRevision`, input.Enabled, userID, id, expectedRevision).Scan(&revision)
	if errors.Is(err, sql.ErrNoRows) {
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM ai_provider_profiles WHERE userID=? AND id=?)`, userID, id).Scan(&exists); err != nil {
			return aiProviderProfile{}, err
		}
		if exists {
			return aiProviderProfile{}, errAIProviderConflict
		}
		return aiProviderProfile{}, errAIProviderNotFound
	}
	if err != nil {
		return aiProviderProfile{}, err
	}
	var oldURL string
	var oldSecret *string
	if err := tx.QueryRowContext(ctx, `SELECT baseURL, secretCiphertext FROM ai_provider_versions WHERE userID=? AND profileID=? AND revision=?`, userID, id, expectedRevision).Scan(&oldURL, &oldSecret); err != nil {
		return aiProviderProfile{}, err
	}
	ciphertext, err := d.resolveAIProviderSecret(input, oldURL, oldSecret)
	if err != nil {
		return aiProviderProfile{}, err
	}
	if input.Secret != nil && *input.Secret == "" {
		if _, err := tx.ExecContext(ctx, `UPDATE ai_provider_versions SET secretCiphertext=NULL WHERE userID=? AND profileID=?`, userID, id); err != nil {
			return aiProviderProfile{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO ai_provider_versions (userID, profileID, revision, name, baseURL, model, secretCiphertext) VALUES (?, ?, ?, ?, ?, ?, ?)`, userID, id, revision, input.Name, input.BaseURL, input.Model, ciphertext); err != nil {
		return aiProviderProfile{}, err
	}
	if err := tx.Commit(); err != nil {
		return aiProviderProfile{}, err
	}
	return aiProviderProfile{Config: input.Config, ID: id, Revision: revision, Enabled: input.Enabled, HasSecret: ciphertext != nil}, nil
}

func (d *Database) resolveAIProviderSecret(input providers.Input, oldURL string, oldSecret *string) (*string, error) {
	if input.Secret == nil {
		if oldSecret == nil {
			return nil, nil
		}
		if input.BaseURL != oldURL {
			return nil, &aiProviderInputError{"changing the base URL requires explicit secret replacement or removal"}
		}
		if !strings.HasPrefix(*oldSecret, encryptedPrefix) {
			return nil, errors.New("stored AI credential is not encrypted")
		}
		if _, err := decryptValue(d.encryptionKey, *oldSecret); err != nil {
			return nil, errors.New("stored AI credential cannot be decrypted")
		}
		return oldSecret, nil
	}
	if *input.Secret == "" {
		return nil, nil
	}
	ciphertext, err := encryptValue(d.encryptionKey, *input.Secret)
	if err != nil {
		return nil, err
	}
	return &ciphertext, nil
}
