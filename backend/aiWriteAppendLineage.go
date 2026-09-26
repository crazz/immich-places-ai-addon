package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func aiReadAppendLineage(ctx context.Context, tx *sql.Tx, owner, installation, asset string) (*writepreview.AppendLineage, error) {
	var lineage writepreview.AppendLineage
	var count int
	err := tx.QueryRowContext(ctx, `WITH owned AS (
 SELECT l.lineageID,l.block,l.hash FROM ai_write_append_lineages l JOIN ai_standard_write_fields f ON f.userID=l.userID AND f.installationID=l.installationID AND f.operationID=l.operationID AND f.field='description' AND f.wasVerified=1 WHERE l.userID=? AND l.installationID=? AND l.assetID=?
 UNION ALL
 SELECT l.lineageID,l.block,l.hash FROM ai_stack_append_lineages l JOIN ai_stack_write_fields f ON f.userID=l.userID AND f.installationID=l.installationID AND f.operationID=l.operationID AND f.assetID=l.assetID AND f.field='description' AND f.wasVerified=1 WHERE l.userID=? AND l.installationID=? AND l.assetID=?
 ) SELECT lineageID,block,hash,(SELECT count(*) FROM owned) FROM owned`, owner, installation, asset, owner, installation, asset).Scan(&lineage.ID, &lineage.Block, &lineage.Hash, &count)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil || count != 1 || len(lineage.Block) > 20<<10 {
		return nil, drafts.ErrStorage
	}
	sum := sha256.Sum256([]byte(lineage.Block))
	if _, err = uuid.Parse(lineage.ID); err != nil || lineage.Hash != hex.EncodeToString(sum[:]) {
		return nil, drafts.ErrStorage
	}
	return &lineage, nil
}

func aiRecordAppendLineage(ctx context.Context, tx *sql.Tx, op writeback.Operation, fields []writeback.FieldOutcome) error {
	if op.Plan.Description == nil {
		return nil
	}
	verified := false
	for _, field := range fields {
		verified = verified || (field.Field == "description" && field.Status == "verified")
	}
	if !verified {
		return nil
	}
	table, previous := "ai_write_append_lineages", "ai_stack_append_lineages"
	if op.Plan.Version == "stack-preview-v3" {
		table, previous = previous, table
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM `+previous+` WHERE userID=? AND installationID=? AND assetID=?`, op.Plan.Owner, op.Plan.Installation, op.Plan.TargetID); err != nil {
		return drafts.ErrStorage
	}
	lineage := op.Plan.Description.Lineage
	if lineage == nil {
		if _, err := tx.ExecContext(ctx, `DELETE FROM `+table+` WHERE userID=? AND installationID=? AND assetID=?`, op.Plan.Owner, op.Plan.Installation, op.Plan.TargetID); err != nil {
			return drafts.ErrStorage
		}
		return nil
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO `+table+`(userID,installationID,assetID,operationID,lineageID,block,hash) VALUES(?,?,?,?,?,?,?) ON CONFLICT(userID,installationID,assetID) DO UPDATE SET operationID=excluded.operationID,lineageID=excluded.lineageID,block=excluded.block,hash=excluded.hash`, op.Plan.Owner, op.Plan.Installation, op.Plan.TargetID, op.ID, lineage.ID, lineage.Block, lineage.Hash)
	if err != nil {
		return drafts.ErrStorage
	}
	return nil
}
