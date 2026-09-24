package main

import (
	"context"
	"database/sql"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/results"
)

func (s *aiDraftStore) edit(ctx context.Context, owner, id string, revision int, edit drafts.Edit) (drafts.Draft, error) {
	var value drafts.Draft
	var document *results.Document
	if edit.CandidateID != nil {
		current, err := s.get(ctx, owner, id)
		if err != nil {
			return value, err
		}
		detail, err := s.results.detail(ctx, owner, current.AnalysisID, "", "")
		if err != nil {
			return value, drafts.ErrUnavailable
		}
		document = detail.Proposal
	}
	err := s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, err := s.read(ctx, tx, owner, id)
		if err != nil {
			return err
		}
		if current.Revision != revision {
			return drafts.ErrConflict
		}
		value, err = drafts.Apply(current, edit, document)
		if err != nil {
			return err
		}
		value.UpdatedAt = s.results.jobs.now().UTC().Format(time.RFC3339Nano)
		if err = s.snapshot(ctx, tx, owner, value); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE ai_drafts SET revision=?,state=? WHERE userID=? AND installationID=? AND id=?`, value.Revision, value.State, owner, s.results.jobs.binding, id)
		if err != nil {
			return drafts.ErrStorage
		}
		return nil
	})
	return value, err
}
