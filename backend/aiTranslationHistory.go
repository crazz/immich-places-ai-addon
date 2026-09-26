package main

import (
	"context"
	"database/sql"
	"immich-places-backend/internal/ai/drafts"
)

type aiTranslationPage struct {
	IDs  []string `json:"ids"`
	Next string   `json:"next"`
}

func (s *aiTranslationStore) list(ctx context.Context, owner, draft, before string) (aiTranslationPage, error) {
	page := aiTranslationPage{IDs: []string{}}
	err := s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := s.drafts.read(ctx, tx, owner, draft); err != nil {
			return err
		}
		binding := s.drafts.results.jobs.binding
		var at int64
		if before != "" {
			if tx.QueryRowContext(ctx, `SELECT createdAt FROM ai_translation_runs WHERE userID=? AND installationID=? AND draftID=? AND id=?`, owner, binding, draft, before).Scan(&at) != nil {
				return drafts.ErrUnavailable
			}
		}
		rows, err := tx.QueryContext(ctx, `SELECT id FROM ai_translation_runs WHERE userID=? AND installationID=? AND draftID=? AND (?='' OR (createdAt,id)<(?,?)) ORDER BY createdAt DESC,id DESC LIMIT 21`, owner, binding, draft, before, at, before)
		if err != nil {
			return drafts.ErrStorage
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			if rows.Scan(&id) != nil {
				return drafts.ErrStorage
			}
			page.IDs = append(page.IDs, id)
		}
		if rows.Err() != nil {
			return drafts.ErrStorage
		}
		if len(page.IDs) > 20 {
			page.IDs = page.IDs[:20]
			page.Next = page.IDs[19]
		}
		return nil
	})
	return page, err
}
