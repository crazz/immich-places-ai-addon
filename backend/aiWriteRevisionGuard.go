package main

import (
	"context"
	"database/sql"

	"immich-places-backend/internal/ai/drafts"
)

func (s *aiDraftStore) guardWriteRevision(ctx context.Context, tx *sql.Tx, owner, id string) error {
	binding := s.results.jobs.binding
	var active int
	err := tx.QueryRowContext(ctx, `SELECT count(*) FROM ai_write_operations o JOIN ai_write_targets t ON t.userID=o.userID AND t.installationID=o.installationID AND t.operationID=o.id WHERE o.userID=? AND o.installationID=? AND o.draftID=? AND (o.status IN ('writing','verifying') OR t.senderActive=1 OR (t.attempts>0 AND t.completionKnown=0))`, owner, binding, id).Scan(&active)
	if err != nil {
		return drafts.ErrStorage
	}
	if active > 0 {
		return drafts.ErrWriteInProgress
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO ai_write_events(userID,installationID,operationID,code,at,attempt) SELECT o.userID,o.installationID,o.id,'DRAFT_CHANGED',?,t.attempts FROM ai_write_operations o JOIN ai_write_targets t ON t.userID=o.userID AND t.installationID=o.installationID AND t.operationID=o.id WHERE o.userID=? AND o.installationID=? AND o.draftID=? AND o.status IN ('queued','retryable')`, s.results.jobs.now().UnixNano(), owner, binding, id); err != nil {
		return drafts.ErrStorage
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM ai_write_target_guards WHERE token IN (SELECT id FROM ai_write_operations WHERE userID=? AND installationID=? AND draftID=? AND status IN ('queued','retryable'))`, owner, binding, id); err != nil {
		return drafts.ErrStorage
	}
	if _, err = tx.ExecContext(ctx, `UPDATE ai_write_operations SET status='canceled',code='DRAFT_CHANGED' WHERE userID=? AND installationID=? AND draftID=? AND status IN ('queued','retryable')`, owner, binding, id); err != nil {
		return drafts.ErrStorage
	}
	if _, err = tx.ExecContext(ctx, `UPDATE ai_write_previews SET invalidated=1 WHERE userID=? AND installationID=? AND draftID=?`, owner, binding, id); err != nil {
		return drafts.ErrStorage
	}
	return nil
}
