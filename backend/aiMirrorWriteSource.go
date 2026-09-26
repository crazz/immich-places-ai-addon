package main

import (
	"context"
	"database/sql"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func (a *aiMirrorAttempt) read(ctx context.Context, allowMissing bool) (writepreview.Metadata, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	p, s := a.standard.parent, a.standard.store
	fresh, _, authority, err := a.standard.readAsset(ctx, p.Plan.TargetID, allowMissing)
	if err != nil {
		return fresh, err
	}
	raw, _, err := s.images.fetch(ctx, authority.key, "/api/assets/"+p.Plan.TargetID+"/metadata", writepreview.MirrorListLimit)
	if err != nil {
		return fresh, err
	}
	defer clear(raw)
	baseline, err := writepreview.ParseMirrorNamespace(raw)
	if err != nil {
		return fresh, err
	}
	after, _, afterAuthority, err := a.standard.readAsset(ctx, p.Plan.TargetID, allowMissing)
	if err != nil || afterAuthority != authority || after.ImageIdentity != fresh.ImageIdentity {
		return fresh, writeback.Failure("SOURCE_CHANGED")
	}
	err = s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, err := a.standard.assetAuthority(ctx, tx, p.Plan.TargetID, allowMissing)
		if err != nil || current != authority {
			return drafts.ErrUnavailable
		}
		return nil
	})
	if err != nil {
		return fresh, err
	}
	a.standard.authority, a.standard.analyzedAuthority = authority, authority
	after.Mirror = &baseline
	return after, nil
}
