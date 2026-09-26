package main

import (
	"context"
	"database/sql"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/translations"
	"time"
)

func (s *aiTranslationStore) adopt(ctx context.Context, owner, id string, revision int, languages []string) (drafts.Draft, error) {
	var value drafts.Draft
	err := s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		run, err := s.read(ctx, tx, owner, id)
		if err != nil {
			return err
		}
		current, err := s.drafts.read(ctx, tx, owner, run.Request.DraftID)
		if err != nil {
			return err
		}
		selected := []translations.Outcome{}
		for _, tag := range languages {
			for _, item := range run.Items {
				if tag == item.Language {
					selected = append(selected, translations.Outcome{Language: tag, Status: item.State, Text: item.Text})
				}
			}
		}
		if len(selected) != len(languages) {
			return drafts.ErrInvalid
		}
		value, err = translations.Adopt(current, run.Request, revision, selected)
		if err != nil {
			return err
		}
		value.UpdatedAt = s.drafts.results.jobs.now().UTC().Format(time.RFC3339Nano)
		if err = s.drafts.snapshot(ctx, tx, owner, value); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE ai_drafts SET revision=?,state=? WHERE userID=? AND installationID=? AND id=?`, value.Revision, value.State, owner, s.drafts.results.jobs.binding, value.ID)
		return err
	})
	return value, err
}
