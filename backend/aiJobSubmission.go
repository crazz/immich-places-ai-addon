package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/jobs"
)

func (s *aiJobStore) Submit(ctx context.Context, input jobs.Submission) (jobs.Job, error) {
	input, digest, err := jobs.Normalize(input)
	if err != nil {
		return jobs.Job{}, err
	}
	data, err := json.Marshal(input)
	if err != nil {
		return jobs.Job{}, jobs.ErrInvalid
	}
	id := uuid.NewString()
	err = s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := s.current(ctx, tx); err != nil {
			return err
		}
		if input.Installation != s.binding {
			return jobs.ErrDenied
		}
		var model string
		if err := tx.QueryRowContext(ctx, `SELECT v.model FROM ai_provider_profiles p JOIN ai_provider_versions v ON v.userID=p.userID AND v.profileID=p.id WHERE p.userID=? AND p.id=? AND v.revision=? AND p.enabled=1`, input.Owner, input.Profile, input.Revision).Scan(&model); err != nil {
			return jobs.ErrDenied
		}
		var existingID, existingDigest string
		err := tx.QueryRowContext(ctx, `SELECT id,requestDigest FROM ai_jobs WHERE userID=? AND installationID=? AND idempotencyKey=?`, input.Owner, input.Installation, input.Key).Scan(&existingID, &existingDigest)
		if err == nil {
			if existingDigest != digest {
				return jobs.ErrConflict
			}
			id = existingID
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO ai_jobs(userID,id,installationID,idempotencyKey,requestDigest,requestJSON,model,maxCalls,createdAt) VALUES(?,?,?,?,?,?,?,?,?)`, input.Owner, id, input.Installation, input.Key, digest, string(data), model, input.MaxCalls, s.now().UnixNano()); err != nil {
			return err
		}
		for position, asset := range input.AssetIDs {
			if _, err := tx.ExecContext(ctx, `INSERT INTO ai_job_items(userID,jobID,id,assetID,position) VALUES(?,?,?,?,?)`, input.Owner, id, uuid.NewString(), asset, position); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return jobs.Job{}, err
	}
	return s.Get(ctx, input.Owner, id)
}
