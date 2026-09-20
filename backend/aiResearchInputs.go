package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"immich-places-backend/internal/ai/contextual"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/selection"
)

type aiResearchInputs struct {
	CaptureTime string
	Album       *contextual.Album
}

func loadAIResearchInputs(ctx context.Context, tx *sql.Tx, lease jobs.Lease, cfg jobs.Configuration) (*aiResearchInputs, error) {
	var raw string
	if err := tx.QueryRowContext(ctx, `SELECT selectionJSON FROM ai_job_admissions WHERE userID=? AND jobID=?`, lease.Owner, lease.JobID).Scan(&raw); err != nil {
		return nil, err
	}
	var manifest selection.Manifest
	if json.Unmarshal([]byte(raw), &manifest) != nil {
		return nil, jobs.ErrDenied
	}
	input := &aiResearchInputs{}
	if manifest.ContextPreview == nil {
		return input, nil
	}
	input.CaptureTime = manifest.ContextPreview.CaptureTimes[lease.Asset]
	if cfg.Context.AlbumID != "" && cfg.Context.AlbumID == manifest.Scope.AlbumID && manifest.ContextPreview.AlbumLabel != nil {
		input.Album = &contextual.Album{ID: cfg.Context.AlbumID, Label: *manifest.ContextPreview.AlbumLabel, Member: true}
	}
	return input, nil
}
