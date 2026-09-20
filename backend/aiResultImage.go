package main

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"immich-places-backend/internal/ai/images"
	"immich-places-backend/internal/ai/review"
)

func (s *aiResultStore) imageAuthority(ctx context.Context, owner, asset string) (aiImageAuthority, error) {
	var binding string
	if s.jobs.db.db.QueryRowContext(ctx, `SELECT id FROM ai_installation_identity WHERE singleton=1`).Scan(&binding) != nil || binding != s.jobs.binding {
		return aiImageAuthority{}, review.ErrUnavailable
	}
	user, err := s.jobs.db.getUserByID(ctx, owner)
	if err != nil || user == nil || user.ImmichAPIKey == nil || strings.TrimSpace(*user.ImmichAPIKey) == "" {
		return aiImageAuthority{}, review.ErrUnavailable
	}
	row, err := s.jobs.db.getAssetByID(ctx, owner, asset)
	if err != nil || row == nil || row.Type != "IMAGE" || row.IsHidden {
		return aiImageAuthority{}, review.ErrUnavailable
	}
	facts, err := json.Marshal(struct {
		Type                    string
		Hidden                  bool
		Library, Stack, Primary *string
	}{row.Type, row.IsHidden, row.LibraryID, row.StackID, row.StackPrimaryAssetID})
	if err != nil {
		return aiImageAuthority{}, review.ErrUnavailable
	}
	return aiImageAuthority{key: *user.ImmichAPIKey, facts: string(facts)}, nil
}

func (s *aiResultStore) thumbnail(ctx context.Context, owner, job, item string, reader *aiImagePreparer) ([]byte, error) {
	if reader == nil {
		return nil, review.ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	detail, err := s.detail(ctx, owner, "", job, item)
	if err != nil {
		return nil, err
	}
	asset := detail.Entry.AssetID
	authority, err := s.imageAuthority(ctx, owner, asset)
	if err != nil {
		return nil, err
	}
	source, err := reader.sourceDigest(ctx, authority.key, asset)
	if err != nil {
		return nil, review.ErrUnavailable
	}
	if current, err := s.imageAuthority(ctx, owner, asset); err != nil || current != authority {
		return nil, review.ErrUnavailable
	}
	data, mime, err := reader.fetch(ctx, authority.key, "/api/assets/"+asset+"/thumbnail?size=preview", reader.limits.MaxSourceBytes)
	if err != nil {
		return nil, review.ErrUnavailable
	}
	defer clear(data)
	prepared, err := images.PrepareRaster(ctx, data, mime, images.Binding{Owner: owner, Installation: s.jobs.binding, Asset: asset, SourceDigest: source}, reader.limits)
	if err != nil {
		return nil, review.ErrUnavailable
	}
	defer prepared.Release()
	if current, err := s.imageAuthority(ctx, owner, asset); err != nil || current != authority {
		return nil, review.ErrUnavailable
	}
	if current, err := reader.sourceDigest(ctx, authority.key, asset); err != nil || current != source {
		return nil, review.ErrUnavailable
	}
	if current, err := s.imageAuthority(ctx, owner, asset); err != nil || current != authority || ctx.Err() != nil {
		return nil, review.ErrUnavailable
	}
	return prepared.Bytes()
}
