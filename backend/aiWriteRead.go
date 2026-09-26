package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func (s *aiWriteStore) get(ctx context.Context, owner, id string, byKey bool) (writeback.Operation, error) {
	var op writeback.Operation
	err := s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		op, err = s.read(ctx, tx, owner, id, byKey)
		return err
	})
	return op, err
}

func (s *aiWriteStore) read(ctx context.Context, tx *sql.Tx, owner, id string, byKey bool) (writeback.Operation, error) {
	stack, err := s.readStackOperation(ctx, tx, owner, id, byKey)
	if !errors.Is(err, sql.ErrNoRows) {
		return stack, err
	}
	var op writeback.Operation
	var raw, observed []byte
	var at int64
	var previewID, draftID, assetID string
	var revision int
	column := "o.id"
	if byKey {
		column = "o.idempotencyKey"
	}
	err = tx.QueryRowContext(ctx, `SELECT o.id,o.payload,o.digest,o.status,o.code,o.approvedAt,o.previewID,o.draftID,o.revision,o.assetID,t.attempts,t.generation,t.observed,t.verified,t.refreshed,t.noop,(t.completionKnown=1 AND t.senderActive=0) FROM ai_all_write_operations o JOIN ai_all_write_targets t ON t.userID=o.userID AND t.installationID=o.installationID AND t.operationID=o.id WHERE o.userID=? AND o.installationID=? AND `+column+`=?`, owner, s.drafts.results.jobs.binding, id).Scan(&op.ID, &raw, &op.Digest, &op.Status, &op.Code, &at, &previewID, &draftID, &revision, &assetID, &op.Attempts, &op.Generation, &observed, &op.Verified, &op.Refreshed, &op.Noop, &op.Settled)
	if errors.Is(err, sql.ErrNoRows) {
		return op, drafts.ErrUnavailable
	}
	if err != nil {
		return op, drafts.ErrStorage
	}
	op.Plan, err = writeback.DecodePlan(raw, op.Digest, owner, s.drafts.results.jobs.binding, previewID)
	if err != nil || op.Plan.DraftID != draftID || op.Plan.DraftRevision != revision || op.Plan.TargetID != assetID {
		return op, drafts.ErrStorage
	}
	if len(observed) > 0 && json.Unmarshal(observed, &op.Observed) != nil {
		return op, drafts.ErrStorage
	}
	op.ApprovedAt = time.Unix(0, at).UTC().Format(time.RFC3339Nano)
	if err = aiReadWriteFields(ctx, tx, &op); err != nil {
		return op, err
	}
	op.Events = []writeback.Event{}
	rows, err := tx.QueryContext(ctx, aiWriteStatement(op.Plan, `SELECT code,at,attempt FROM ai_write_events WHERE userID=? AND installationID=? AND operationID=? ORDER BY sequence LIMIT 100`), owner, op.Plan.Installation, op.ID)
	if err != nil {
		return op, drafts.ErrStorage
	}
	defer rows.Close()
	for rows.Next() {
		var e writeback.Event
		if err = rows.Scan(&e.Code, &at, &e.Attempt); err != nil {
			return op, drafts.ErrStorage
		}
		e.At = time.Unix(0, at).UTC().Format(time.RFC3339Nano)
		op.Events = append(op.Events, e)
	}
	return op, rows.Err()
}
