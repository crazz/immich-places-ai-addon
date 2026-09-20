package main

import (
	"context"
	"immich-places-backend/internal/ai/contextual"
	"time"
)

type aiContextRequest struct {
	Binding   contextual.Binding
	Consent   contextual.Consent
	Hint      string
	Window    time.Duration
	Authorize func(context.Context) error
	Displayed *aiResearchInputs
}
type aiContextPreparer struct {
	images  *aiImagePreparer
	lineage func(context.Context, string, string) (string, error)
}

func (s *aiContextPreparer) prepare(ctx context.Context, req aiContextRequest) (*contextual.Bundle, error) {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	if s == nil || s.images == nil || req.Authorize == nil || contextual.ValidateConsent(req.Binding, req.Consent) != nil {
		return nil, contextual.ErrInvalid
	}
	if err := req.Authorize(ctx); err != nil {
		return nil, errAIImageDenied
	}
	authority, err := s.images.authorize(ctx, req.Binding.Owner, req.Binding.Installation, req.Binding.Asset)
	if err != nil {
		return nil, errAIImageDenied
	}
	data, _, err := s.images.fetch(ctx, authority.key, "/api/assets/"+req.Binding.Asset, 1<<20)
	if err != nil {
		return nil, err
	}
	defer clear(data)
	digest, err := aiImageSourceDigest(data, req.Binding.Asset)
	if err != nil || digest != req.Binding.SourceDigest {
		return nil, errAIImageDenied
	}
	in := contextual.Input{Binding: req.Binding, Consent: req.Consent, Hint: req.Hint, Window: req.Window}
	if req.Displayed != nil {
		in.CaptureTime, in.Album = req.Displayed.CaptureTime, req.Displayed.Album
	}
	if req.Displayed == nil && (req.Consent.Has(contextual.Capture) || req.Consent.Has(contextual.Neighbors)) {
		exif, err := aiContextExif(data)
		if err != nil {
			return nil, err
		}
		if exif != nil && exif.DateTimeOriginal != nil {
			in.CaptureTime = *exif.DateTimeOriginal
		}
	}
	if req.Consent.Has(contextual.Neighbors) {
		in.Candidates, err = s.neighbors(ctx, req, in.CaptureTime)
		if err != nil {
			return nil, err
		}
	}
	if req.Displayed == nil && req.Consent.Has(contextual.AlbumLabel) {
		in.Album, err = s.album(ctx, req, authority.key)
		if err != nil {
			return nil, err
		}
	}
	if current, err := s.images.authorize(ctx, req.Binding.Owner, req.Binding.Installation, req.Binding.Asset); err != nil || current != authority {
		return nil, errAIImageDenied
	}
	if err := req.Authorize(ctx); err != nil {
		return nil, errAIImageDenied
	}
	return contextual.Build(in)
}
