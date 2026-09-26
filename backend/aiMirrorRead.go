package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func aiReadMirrorStep(ctx context.Context, tx *sql.Tx, op writeback.Operation) (*writeback.MirrorOutcome, error) {
	m := &writeback.MirrorOutcome{Step: "metadata", Events: []writeback.MirrorEvent{}}
	var raw []byte
	err := tx.QueryRowContext(ctx, `SELECT assetID,status,code,attempts,generation,observed,verified,noop,(completionKnown=1 AND senderActive=0) FROM ai_mirror_write_steps WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, op.Plan.Owner, op.Plan.Installation, op.ID, op.Plan.TargetID).Scan(&m.AssetID, &m.Status, &m.Code, &m.Attempts, &m.Generation, &raw, &m.Verified, &m.Noop, &m.Settled)
	if err != nil || len(raw) > 128<<10 || (len(raw) > 0 && json.Unmarshal(raw, &m.Observed) != nil) {
		return nil, drafts.ErrStorage
	}
	rows, err := tx.QueryContext(ctx, `SELECT code,at,attempt,generation FROM ai_mirror_write_events WHERE userID=? AND installationID=? AND operationID=? AND assetID=? ORDER BY sequence LIMIT 100`, op.Plan.Owner, op.Plan.Installation, op.ID, op.Plan.TargetID)
	if err != nil {
		return nil, drafts.ErrStorage
	}
	defer rows.Close()
	for rows.Next() {
		var event writeback.MirrorEvent
		var at int64
		if rows.Scan(&event.Code, &at, &event.Attempt, &event.Generation) != nil {
			return nil, drafts.ErrStorage
		}
		event.At = time.Unix(0, at).UTC().Format(time.RFC3339Nano)
		m.Events = append(m.Events, event)
	}
	return m, rows.Err()
}
