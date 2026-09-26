package main

import (
	"context"
	"immich-places-backend/internal/ai/providers"
	"time"
)

type aiTranslationRuntime struct {
	store      *aiTranslationStore
	dispatcher *providers.Dispatcher
}

func (r *aiTranslationRuntime) run(ctx context.Context, onFailure func()) error {
	if err := r.store.recover(ctx); err != nil {
		return err
	}
	if !r.store.drafts.results.jobs.enabled {
		return nil
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for ctx.Err() == nil {
		worked, err := r.store.runOne(ctx, r.dispatcher)
		if ctx.Err() != nil {
			return nil
		}
		if err != nil && onFailure != nil {
			onFailure()
		}
		if worked && err == nil {
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
	return nil
}
