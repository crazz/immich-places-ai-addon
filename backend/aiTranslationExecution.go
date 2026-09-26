package main

import (
	"context"
	"database/sql"
	"errors"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/ai/translations"
)

func (s *aiTranslationStore) runOne(ctx context.Context, dispatcher *providers.Dispatcher) (bool, error) {
	return (translations.Worker{Store: aiTranslationWorkerStore{store: s}, Provider: aiTextTranslator{dispatcher: dispatcher}}).RunOne(ctx)
}
func (w aiTranslationWorkerStore) Claim(ctx context.Context) (translations.Task, bool, error) {
	s := w.store
	var owner, id, tag string
	var run aiTranslationRun
	capacity := s.capacity
	if !capacity.Valid() {
		capacity = jobs.DefaultPolicy()
	}
	err := s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		now := s.drafts.results.jobs.now().UnixNano()
		var active int
		if tx.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM ai_job_items WHERE state='running' AND leaseExpiresAt>?) + (SELECT count(*) FROM ai_translation_items WHERE state='reserved')`, now).Scan(&active) != nil {
			return drafts.ErrStorage
		}
		if active >= capacity.Global {
			return nil
		}
		err := tx.QueryRowContext(ctx, `SELECT i.userID,i.runID,i.language FROM ai_translation_items i JOIN ai_translation_runs r ON r.userID=i.userID AND r.installationID=i.installationID AND r.id=i.runID WHERE i.installationID=? AND i.state='queued'
 AND ((SELECT count(*) FROM ai_job_items a WHERE a.userID=i.userID AND a.state='running' AND a.leaseExpiresAt>?) + (SELECT count(*) FROM ai_translation_items t WHERE t.userID=i.userID AND t.state='reserved'))<?
 ORDER BY r.createdAt,r.id,i.language LIMIT 1`, s.drafts.results.jobs.binding, now, capacity.PerOwner).Scan(&owner, &id, &tag)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return drafts.ErrStorage
		}
		run, err = s.read(ctx, tx, owner, id)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE ai_translation_items SET state='reserved' WHERE userID=? AND installationID=? AND runID=? AND language=? AND state='queued'`, owner, s.drafts.results.jobs.binding, id, tag)
		return err
	})
	if err != nil || id == "" {
		return translations.Task{}, false, err
	}
	return translations.Task{Owner: owner, ID: id, Language: tag, PolicyID: run.PolicyID, Request: run.Request, Policy: run.Policy}, true, nil
}

func (s *aiTranslationStore) finish(ctx context.Context, owner, id, tag, state string, text *string, failure string) error {
	if ctx.Err() != nil {
		state, text, failure = "interrupted", nil, "process_interrupted"
	}
	return s.drafts.write(context.WithoutCancel(ctx), func(ctx context.Context, tx *sql.Tx) error {
		run, err := s.read(ctx, tx, owner, id)
		if err != nil {
			return err
		}
		if err = s.authorizeItem(ctx, tx, owner, run, tag); err != nil {
			state, text, failure = "interrupted", nil, "authority_changed"
		}
		var canceled bool
		if tx.QueryRowContext(ctx, `SELECT cancelRequested FROM ai_translation_items WHERE userID=? AND installationID=? AND runID=? AND language=?`, owner, s.drafts.results.jobs.binding, id, tag).Scan(&canceled) != nil {
			return drafts.ErrStorage
		}
		if canceled {
			state, text, failure = "canceled", nil, ""
		}
		_, err = tx.ExecContext(ctx, `UPDATE ai_translation_items SET state=?,text=?,failure=? WHERE userID=? AND installationID=? AND runID=? AND language=? AND state='reserved'`, state, text, failure, owner, s.drafts.results.jobs.binding, id, tag)
		return err
	})
}
