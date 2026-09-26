package main

import (
	"context"
	"database/sql"
	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/translations"
)

type aiTranslationWorkerStore struct{ store *aiTranslationStore }

func translationTaskRun(t translations.Task) aiTranslationRun {
	return aiTranslationRun{ID: t.ID, Request: t.Request, PolicyID: t.PolicyID, Policy: t.Policy}
}
func (s aiTranslationWorkerStore) Authorize(ctx context.Context, t translations.Task) error {
	return s.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return s.store.authorizeItem(ctx, tx, t.Owner, translationTaskRun(t), t.Language)
	})
}
func (s aiTranslationWorkerStore) Reserve(ctx context.Context, t translations.Task) error {
	return s.store.reserveUsage(ctx, t.Owner, translationTaskRun(t), t.Language)
}
func (s aiTranslationWorkerStore) RecordUsage(ctx context.Context, t translations.Task, usage *analysis.Usage) error {
	return s.store.recordUsage(ctx, t.Owner, translationTaskRun(t), t.Language, usage)
}
func (s aiTranslationWorkerStore) Finish(ctx context.Context, t translations.Task, out translations.Outcome, failure string) error {
	return s.store.finish(ctx, t.Owner, t.ID, t.Language, out.Status, out.Text, failure)
}
