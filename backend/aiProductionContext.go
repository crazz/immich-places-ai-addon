package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"immich-places-backend/internal/ai/contextual"
	"immich-places-backend/internal/ai/images"
	"immich-places-backend/internal/ai/jobs"
)

func (p *aiProductionJobs) frozenContext(ctx context.Context, lease jobs.Lease, cfg jobs.Configuration, image *images.Prepared, preparer *aiContextPreparer, authorize func(context.Context) error) (*contextual.Bundle, aiContextRequest, error) {
	info, ok := image.Info()
	if !ok || cfg.Context == nil {
		return nil, aiContextRequest{}, jobs.ErrDenied
	}
	binding := contextual.Binding{Owner: lease.Owner, Installation: lease.Installation, Asset: lease.Asset, Selection: lease.Input.SelectionDigest, Profile: cfg.ProfileID, Revision: cfg.Revision, SourceDigest: info.Binding.SourceDigest}
	req := aiContextRequest{Binding: binding, Consent: contextual.Consent{Binding: binding, Version: cfg.Context.Version, Classes: cfg.Context.Classes, AlbumID: cfg.Context.AlbumID}, Hint: cfg.Context.Hint, Window: 6 * time.Hour, Authorize: authorize}
	var bundle *contextual.Bundle
	err := p.store.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := p.store.leaseState(ctx, tx, lease); err != nil {
			return err
		}
		if cfg.Mode == "research" {
			var err error
			req.Displayed, err = loadAIResearchInputs(ctx, tx, lease, cfg)
			if err != nil {
				return err
			}
		}
		var raw string
		err := tx.QueryRowContext(ctx, "SELECT bundleJSON FROM ai_job_context WHERE userID=? AND jobID=? AND itemID=?", lease.Owner, lease.JobID, lease.ItemID).Scan(&raw)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		bundle = &contextual.Bundle{}
		if json.Unmarshal([]byte(raw), bundle) != nil || bundle.Info().Binding != binding {
			return jobs.ErrDenied
		}
		return nil
	})
	if err != nil {
		return nil, req, err
	}
	if bundle != nil {
		if err = preparer.check(ctx, req, bundle); err != nil {
			return nil, req, err
		}
		return bundle, req, nil
	}
	bundle, err = preparer.prepare(ctx, req)
	if err != nil {
		return nil, req, err
	}
	raw, err := json.Marshal(bundle)
	if err != nil {
		return nil, req, jobs.ErrInvalid
	}
	err = p.store.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		state, err := p.store.leaseState(ctx, tx, lease)
		if err != nil {
			return err
		}
		if state.Canceled || state.Blocked {
			return jobs.ErrDenied
		}
		if _, _, err = p.executionAuthority(ctx, tx, lease); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO ai_job_context VALUES(?,?,?,?,?)", lease.Owner, lease.JobID, lease.ItemID, string(raw), p.store.now().UnixNano())
		return err
	})
	return bundle, req, err
}
