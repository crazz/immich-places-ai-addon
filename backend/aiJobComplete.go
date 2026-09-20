package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/results"
)

func (s *aiJobStore) Complete(ctx context.Context, lease jobs.Lease, completion jobs.Completion) (string, error) {
	job, err := s.Get(ctx, lease.Owner, lease.JobID)
	if err != nil {
		return "", err
	}
	payload, err := completion.Proposal.MarshalJSON()
	if err != nil {
		return "", jobs.ErrInvalid
	}
	validator, err := results.New()
	if err != nil {
		return "", jobs.ErrStorage
	}
	proposal, err := validator.Validate(payload, results.Context{Mode: results.Visual, Completion: results.Complete, Languages: job.Input.Languages, PrimaryLanguage: job.Input.PrimaryLanguage})
	if err != nil {
		return "", jobs.ErrInvalid
	}
	for _, digest := range []string{completion.SourceDigest, completion.ImageDigest} {
		decoded, err := hex.DecodeString(digest)
		if err != nil || len(decoded) != 32 {
			return "", jobs.ErrInvalid
		}
	}
	if completion.PromptVersion != "visual-v1" || completion.SchemaVersion != "1.0" {
		return "", jobs.ErrInvalid
	}
	document, err := proposal.Data()
	if err != nil {
		return "", jobs.ErrInvalid
	}
	version, err := proposal.PolicyVersion()
	if err != nil {
		return "", jobs.ErrInvalid
	}
	metadata, err := json.Marshal(jobs.ResultMetadata{Installation: job.Input.Installation, Profile: job.Input.Profile, Model: job.Model, Revision: job.Input.Revision, Languages: job.Input.Languages, PrimaryLanguage: job.Input.PrimaryLanguage, SelectionDigest: job.Input.SelectionDigest, SourceDigest: completion.SourceDigest, ImageDigest: completion.ImageDigest, PromptVersion: completion.PromptVersion, SchemaVersion: completion.SchemaVersion, ValidationVersion: version})
	if err != nil {
		return "", jobs.ErrInvalid
	}
	id := uuid.NewString()
	err = s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		state, err := s.leaseState(ctx, tx, lease)
		if err != nil {
			return err
		}
		if state.Canceled || state.Blocked {
			return jobs.ErrDenied
		}
		if state.LeaseCalls == 0 {
			return jobs.ErrBudget
		}
		now := s.now().UnixNano()
		if _, err = tx.ExecContext(ctx, `INSERT INTO ai_analyses(userID,id,jobID,itemID,assetID,outcome,payload,metadata,createdAt) VALUES(?,?,?,?,?,?,?,?,?)`, lease.Owner, id, lease.JobID, lease.ItemID, lease.Asset, document.Outcome, string(payload), string(metadata), now); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE ai_job_items SET state='succeeded',leaseToken=NULL,leaseExpiresAt=NULL,finishedAt=?,failure='' WHERE userID=? AND jobID=? AND id=? AND leaseToken=?`, now, lease.Owner, lease.JobID, lease.ItemID, lease.Token)
		return err
	})
	if err != nil {
		return "", err
	}
	return id, nil
}
