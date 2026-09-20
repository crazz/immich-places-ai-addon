package main

import (
	"context"
	"errors"
	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/contextual"
)

var errAIContextChanged = errors.Join(analysis.ErrDenied, errors.New("AI context changed"))

func (s *aiContextPreparer) check(ctx context.Context, req aiContextRequest, frozen *contextual.Bundle) error {
	if frozen == nil || frozen.Info().Binding != req.Binding {
		return errAIImageDenied
	}
	current, err := s.prepare(ctx, req)
	if err != nil {
		return err
	}
	if current.Info().Digest != frozen.Info().Digest {
		return errAIContextChanged
	}
	return nil
}
