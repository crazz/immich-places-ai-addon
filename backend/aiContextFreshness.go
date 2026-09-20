package main

import (
	"context"
	"immich-places-backend/internal/ai/contextual"
)

func (s *aiContextPreparer) check(ctx context.Context, req aiContextRequest, frozen *contextual.Bundle) error {
	if frozen == nil || frozen.Info().Binding != req.Binding {
		return errAIImageDenied
	}
	current, err := s.prepare(ctx, req)
	if err != nil {
		return err
	}
	if current.Info().Digest != frozen.Info().Digest {
		return errAIImageDenied
	}
	return nil
}
