package main

import (
	"context"
	"database/sql"

	"immich-places-backend/internal/ai/drafts"
)

func (s *aiDraftStore) guardMirrorRevision(ctx context.Context, tx *sql.Tx, owner, id string) error {
	binding := s.results.jobs.binding
	var active int
	if tx.QueryRowContext(ctx, `SELECT count(*) FROM ai_mirror_write_steps m JOIN ai_stack_write_operations o ON o.userID=m.userID AND o.installationID=m.installationID AND o.id=m.operationID WHERE o.userID=? AND o.installationID=? AND o.draftID=? AND (m.status IN ('writing','verifying') OR m.senderActive=1 OR (m.attempts>0 AND m.completionKnown=0))`, owner, binding, id).Scan(&active) != nil {
		return drafts.ErrStorage
	}
	if active > 0 {
		return drafts.ErrWriteInProgress
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO ai_mirror_write_events(userID,installationID,operationID,assetID,code,at,attempt,generation) SELECT m.userID,m.installationID,m.operationID,m.assetID,'DRAFT_CHANGED',?,m.attempts,m.generation FROM ai_mirror_write_steps m JOIN ai_stack_write_operations o ON o.userID=m.userID AND o.installationID=m.installationID AND o.id=m.operationID WHERE o.userID=? AND o.installationID=? AND o.draftID=? AND m.status IN ('blocked','queued','retryable')`, s.results.jobs.now().UnixNano(), owner, binding, id); err != nil {
		return drafts.ErrStorage
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM ai_write_target_guards WHERE token IN (SELECT t.guardToken FROM ai_stack_write_targets t JOIN ai_mirror_write_steps m ON m.userID=t.userID AND m.installationID=t.installationID AND m.operationID=t.operationID AND m.assetID=t.assetID JOIN ai_stack_write_operations o ON o.userID=t.userID AND o.installationID=t.installationID AND o.id=t.operationID WHERE o.userID=? AND o.installationID=? AND o.draftID=? AND m.status IN ('blocked','queued','retryable') AND t.completionKnown=1 AND t.senderActive=0)`, owner, binding, id); err != nil {
		return drafts.ErrStorage
	}
	if _, err := tx.ExecContext(ctx, `UPDATE ai_mirror_write_steps SET status='canceled',code='DRAFT_CHANGED',updatedAt=? WHERE userID=? AND installationID=? AND status IN ('blocked','queued','retryable') AND operationID IN (SELECT id FROM ai_stack_write_operations WHERE userID=? AND installationID=? AND draftID=?)`, s.results.jobs.now().UnixNano(), owner, binding, owner, binding, id); err != nil {
		return drafts.ErrStorage
	}
	return nil
}
