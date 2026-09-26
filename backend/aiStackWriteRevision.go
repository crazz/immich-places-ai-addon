package main

import (
	"context"
	"database/sql"

	"immich-places-backend/internal/ai/drafts"
)

func (s *aiDraftStore) guardStackWriteRevision(ctx context.Context, tx *sql.Tx, owner, id string) error {
	binding := s.results.jobs.binding
	var active int
	if tx.QueryRowContext(ctx, `SELECT count(*) FROM ai_stack_write_operations o JOIN ai_stack_write_targets t ON t.userID=o.userID AND t.installationID=o.installationID AND t.operationID=o.id WHERE o.userID=? AND o.installationID=? AND o.draftID=? AND (t.status IN ('writing','verifying') OR t.senderActive=1 OR (t.attempts>0 AND t.completionKnown=0))`, owner, binding, id).Scan(&active) != nil {
		return drafts.ErrStorage
	}
	if active > 0 {
		return drafts.ErrWriteInProgress
	}
	if err := s.guardMirrorRevision(ctx, tx, owner, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO ai_stack_write_events(userID,installationID,operationID,assetID,code,at,attempt) SELECT t.userID,t.installationID,t.operationID,t.assetID,'DRAFT_CHANGED',?,t.attempts FROM ai_stack_write_targets t JOIN ai_stack_write_operations o ON o.userID=t.userID AND o.installationID=t.installationID AND o.id=t.operationID WHERE o.userID=? AND o.installationID=? AND o.draftID=? AND t.status IN ('queued','retryable')`, s.results.jobs.now().UnixNano(), owner, binding, id); err != nil {
		return drafts.ErrStorage
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM ai_write_target_guards WHERE token IN (SELECT t.guardToken FROM ai_stack_write_targets t JOIN ai_stack_write_operations o ON o.userID=t.userID AND o.installationID=t.installationID AND o.id=t.operationID WHERE o.userID=? AND o.installationID=? AND o.draftID=? AND t.status IN ('queued','retryable'))`, owner, binding, id); err != nil {
		return drafts.ErrStorage
	}
	if _, err := tx.ExecContext(ctx, `UPDATE ai_stack_write_targets SET status='canceled',code='DRAFT_CHANGED' WHERE userID=? AND installationID=? AND status IN ('queued','retryable') AND operationID IN (SELECT id FROM ai_stack_write_operations WHERE userID=? AND installationID=? AND draftID=?)`, owner, binding, owner, binding, id); err != nil {
		return drafts.ErrStorage
	}
	if _, err := tx.ExecContext(ctx, `UPDATE ai_stack_write_previews SET invalidated=1 WHERE userID=? AND installationID=? AND draftID=?`, owner, binding, id); err != nil {
		return drafts.ErrStorage
	}
	return nil
}
