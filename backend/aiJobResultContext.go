package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"immich-places-backend/internal/ai/contextual"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/results"
)

type aiContextRowReader interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func aiJobResultContext(ctx context.Context, db aiContextRowReader, job jobs.Job, lease jobs.Lease, completion jobs.Completion) (results.Context, *contextual.Metadata, error) {
	validation := results.Context{Mode: results.Visual, Completion: results.Complete, Languages: job.Input.Languages, PrimaryLanguage: job.Input.PrimaryLanguage}
	if job.Input.Mode == "visual" || job.Input.Mode == "" {
		if completion.PromptVersion != "visual-v1" {
			return validation, nil, jobs.ErrInvalid
		}
		return validation, nil, nil
	}
	if job.Input.Mode != "context-assisted" || completion.PromptVersion != "context-assisted-v1" {
		return validation, nil, jobs.ErrInvalid
	}
	var raw string
	err := db.QueryRowContext(ctx, "SELECT bundleJSON FROM ai_job_context WHERE userID=? AND jobID=? AND itemID=?", lease.Owner, lease.JobID, lease.ItemID).Scan(&raw)
	if err == sql.ErrNoRows {
		return validation, nil, jobs.ErrDenied
	}
	if err != nil {
		return validation, nil, jobs.ErrStorage
	}
	var bundle contextual.Bundle
	if json.Unmarshal([]byte(raw), &bundle) != nil {
		return validation, nil, jobs.ErrDenied
	}
	m := bundle.Info()
	expected := contextual.Binding{Owner: lease.Owner, Installation: lease.Installation, Asset: lease.Asset, Selection: job.Input.SelectionDigest, Profile: job.Input.Profile, Revision: job.Input.Revision, SourceDigest: completion.SourceDigest}
	if m.Binding != expected {
		return validation, nil, jobs.ErrDenied
	}
	validation.Mode = results.ContextAssisted
	for _, source := range m.Sources {
		validation.Sources = append(validation.Sources, results.Source{ID: source.ID})
	}
	return validation, &m, nil
}
