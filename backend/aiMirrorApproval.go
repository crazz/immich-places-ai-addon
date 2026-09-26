package main

import (
	"context"
	"database/sql"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

func aiApproveMirrorStep(ctx context.Context, tx *sql.Tx, owner string, plan writepreview.Plan, id string, now int64) error {
	var recordID string
	if tx.QueryRowContext(ctx, `SELECT recordID FROM ai_mirror_records WHERE userID=? AND installationID=? AND assetID=?`, owner, plan.Installation, plan.TargetID).Scan(&recordID) != nil || recordID != plan.Mirror.RecordID {
		return drafts.ErrStorage
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO ai_mirror_write_steps(userID,installationID,operationID,assetID,updatedAt) VALUES(?,?,?,?,?)`, owner, plan.Installation, id, plan.TargetID, now); err != nil {
		return drafts.ErrStorage
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO ai_mirror_write_events(userID,installationID,operationID,assetID,code,at,attempt,generation) VALUES(?,?,?,?,'approved',?,0,0)`, owner, plan.Installation, id, plan.TargetID, now)
	if err != nil {
		return drafts.ErrStorage
	}
	return nil
}
