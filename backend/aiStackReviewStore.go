package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

func (s *aiDraftStore) saveStackReview(ctx context.Context, review writepreview.StackReview, sources map[string]aiStackSource) error {
	raw, err := json.Marshal(review)
	expires, parseErr := time.Parse(time.RFC3339Nano, review.ExpiresAt)
	if err != nil || parseErr != nil || len(raw) > 1<<20 {
		return drafts.ErrStorage
	}
	return s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, err := s.read(ctx, tx, review.Owner, review.DraftID)
		if err != nil {
			return err
		}
		now := s.results.jobs.now().UnixNano()
		if current.Revision != review.DraftRevision || current.State != "staged" || expires.UnixNano() <= now {
			return drafts.ErrConflict
		}
		for id, source := range sources {
			if err = s.checkAuthority(ctx, tx, review.Owner, id, source.authority); err != nil {
				return err
			}
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM ai_stack_reviews WHERE rowid IN (SELECT rowid FROM ai_stack_reviews WHERE userID=? AND installationID=? AND expiresAt<=? ORDER BY expiresAt LIMIT 100)`, review.Owner, review.Installation, now); err != nil {
			return drafts.ErrStorage
		}
		var active int
		if tx.QueryRowContext(ctx, `SELECT count(*) FROM ai_stack_reviews WHERE userID=? AND installationID=? AND draftID=? AND expiresAt>?`, review.Owner, review.Installation, review.DraftID, now).Scan(&active) != nil {
			return drafts.ErrStorage
		}
		if active >= 10 {
			return drafts.ErrConflict
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO ai_stack_reviews(userID,installationID,draftID,revision,id,expiresAt,content) VALUES(?,?,?,?,?,?,?)`, review.Owner, review.Installation, review.DraftID, review.DraftRevision, review.ID, expires.UnixNano(), string(raw)); err != nil {
			return drafts.ErrStorage
		}
		return nil
	})
}

func (s *aiDraftStore) readStackReview(ctx context.Context, owner, draft string, revision int, id string) (writepreview.StackReview, error) {
	var review writepreview.StackReview
	err := s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, err := s.read(ctx, tx, owner, draft)
		if err != nil {
			return err
		}
		if current.Revision != revision || current.State != "staged" {
			return drafts.ErrConflict
		}
		var raw []byte
		if tx.QueryRowContext(ctx, `SELECT content FROM ai_stack_reviews WHERE userID=? AND installationID=? AND draftID=? AND revision=? AND id=? AND expiresAt>?`, owner, s.results.jobs.binding, draft, revision, id, s.results.jobs.now().UnixNano()).Scan(&raw) != nil {
			return drafts.ErrUnavailable
		}
		if len(raw) > 1<<20 || json.Unmarshal(raw, &review) != nil {
			return drafts.ErrStorage
		}
		return nil
	})
	return review, err
}
