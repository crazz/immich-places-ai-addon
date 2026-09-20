package main

import (
	"context"
	"time"

	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/providers"
)

type aiProductionRuntime struct {
	jobs     *aiProductionJobs
	consumer jobs.Consumer
	enabled  bool
}

func newAIProductionRuntime(db *Database, cfg *Config, selections *aiSelectionStore, dispatcher *providers.Dispatcher) (*aiProductionRuntime, error) {
	if db == nil || cfg == nil || selections == nil || selections.binding == "" {
		return nil, jobs.ErrInvalid
	}
	p := &aiProductionJobs{store: newAIJobStore(db, selections.binding, cfg.AIEnabled, time.Now), selections: selections, policies: cfg.AIExecutionPolicies, fingerprint: policyFingerprint(cfg.AIProviderEgressPolicy)}
	runtime := &aiProductionRuntime{jobs: p, enabled: cfg.AIEnabled}
	if !cfg.AIEnabled {
		return runtime, nil
	}
	if !cfg.AIJobSettings.Valid() {
		return nil, jobs.ErrInvalid
	}
	images, err := newAIImagePreparer(db, selections, cfg.ImmichURL)
	if err != nil {
		return nil, err
	}
	analyzer, err := newAIVisualAnalyzer(db, images, dispatcher)
	if err != nil {
		return nil, err
	}
	runtime.consumer = jobs.Consumer{Worker: jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Execute: p.executor(analyzer)}, Settings: cfg.AIJobSettings}
	return runtime, nil
}
func (r *aiProductionRuntime) run(ctx context.Context, onFailure func()) error {
	if !r.enabled {
		return nil
	}
	return r.consumer.Run(ctx, onFailure)
}
