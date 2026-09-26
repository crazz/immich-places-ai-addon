package main

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func (s *aiWriteStore) confirmStack(ctx context.Context, tx *sql.Tx, owner string, input writeback.Confirmation, plan writepreview.Plan, raw []byte, digest string) (writeback.Operation, error) {
	var empty writeback.Operation
	key, err := s.credential(ctx, tx, owner)
	if err != nil {
		return empty, err
	}
	id := uuid.NewString()
	now := s.drafts.results.jobs.now().UnixNano()
	_, err = tx.ExecContext(ctx, `INSERT INTO ai_stack_write_operations(userID,installationID,id,previewID,draftID,revision,assetID,idempotencyKey,payload,digest,approvedAt,credentialHash) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, owner, plan.Installation, id, input.PreviewID, plan.DraftID, plan.DraftRevision, plan.TargetID, input.Key, string(raw), digest, now, aiWriteCredentialHash(key))
	if err != nil {
		return empty, drafts.ErrStorage
	}
	for ordinal, target := range plan.Manifest.Targets {
		token := uuid.NewString()
		if _, err = tx.ExecContext(ctx, `INSERT INTO ai_write_target_guards(installationID,assetID,token) VALUES(?,?,?)`, plan.Installation, target.AssetID, token); err != nil {
			return empty, writeback.Failure("TARGET_BUSY")
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO ai_stack_write_targets(userID,installationID,operationID,assetID,ordinal,guardToken) VALUES(?,?,?,?,?,?)`, owner, plan.Installation, id, target.AssetID, ordinal, token); err != nil {
			return empty, drafts.ErrStorage
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO ai_stack_write_events(userID,installationID,operationID,assetID,code,at,attempt) VALUES(?,?,?,?,'approved',?,0)`, owner, plan.Installation, id, target.AssetID, now); err != nil {
			return empty, drafts.ErrStorage
		}
	}
	if plan.Mirror != nil {
		if err = aiApproveMirrorStep(ctx, tx, owner, plan, id, now); err != nil {
			return empty, err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE ai_stack_write_previews SET protected=1 WHERE userID=? AND installationID=? AND id=?`, owner, plan.Installation, input.PreviewID); err != nil {
		return empty, drafts.ErrStorage
	}
	return s.readStackOperation(ctx, tx, owner, id, false)
}
