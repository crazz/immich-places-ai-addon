package main

import (
	"context"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

type aiWriteAttempt struct {
	store     *aiWriteStore
	authority aiImageAuthority
	token     string
}

func (a *aiWriteAttempt) Read(ctx context.Context, op writeback.Operation) (writepreview.Metadata, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var result writepreview.Metadata
	s := a.store
	if s.images == nil {
		return result, drafts.ErrUnavailable
	}
	authority, err := s.readAuthority(ctx, op)
	if err != nil {
		return result, err
	}
	raw, _, err := s.images.fetch(ctx, authority.key, "/api/assets/"+op.Plan.TargetID, 1<<20)
	if err != nil {
		return result, err
	}
	defer clear(raw)
	result, _, err = aiParseWriteMetadata(raw, op.Plan.TargetID, op.Plan.Version == "standard-preview-v2")
	if err != nil {
		return result, err
	}
	current, err := s.readAuthority(ctx, op)
	if err != nil || current != authority {
		return result, drafts.ErrUnavailable
	}
	a.authority = authority
	return result, nil
}
