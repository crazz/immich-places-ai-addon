package main

import (
	"context"
	"database/sql"
)

func (s *aiTranslationStore) cancel(ctx context.Context, owner, id string) error {
	return s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := s.read(ctx, tx, owner, id); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `UPDATE ai_translation_items SET cancelRequested=1,state=CASE WHEN state='queued' THEN 'canceled' ELSE state END WHERE userID=? AND installationID=? AND runID=? AND state IN ('queued','reserved')`, owner, s.drafts.results.jobs.binding, id)
		return err
	})
}

func (s *aiTranslationStore) recover(ctx context.Context) error {
	return s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `UPDATE ai_translation_items SET state=CASE WHEN cancelRequested=1 THEN 'canceled' ELSE 'interrupted' END,failure='process_interrupted' WHERE state='reserved' AND installationID=?`, s.drafts.results.jobs.binding)
		return err
	})
}
