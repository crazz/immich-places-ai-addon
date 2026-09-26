package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func (s *aiWriteStore) readStackOperation(ctx context.Context, tx *sql.Tx, owner, id string, byKey bool) (writeback.Operation, error) {
	var op writeback.Operation
	var raw []byte
	var at int64
	var previewID, draftID, assetID string
	var revision int
	column := "id"
	if byKey {
		column = "idempotencyKey"
	}
	err := tx.QueryRowContext(ctx, `SELECT id,payload,digest,status,code,approvedAt,previewID,draftID,revision,assetID FROM ai_stack_write_operations WHERE userID=? AND installationID=? AND `+column+`=?`, owner, s.drafts.results.jobs.binding, id).Scan(&op.ID, &raw, &op.Digest, &op.Status, &op.Code, &at, &previewID, &draftID, &revision, &assetID)
	if err != nil {
		return op, err
	}
	op.Plan, err = writeback.DecodePlan(raw, op.Digest, owner, s.drafts.results.jobs.binding, previewID)
	if err != nil || op.Plan.Manifest == nil || op.Plan.DraftID != draftID || op.Plan.DraftRevision != revision || op.Plan.TargetID != assetID {
		return op, drafts.ErrStorage
	}
	op.ApprovedAt = time.Unix(0, at).UTC().Format(time.RFC3339Nano)
	op.Events = []writeback.Event{}
	op.Targets, err = aiReadStackTargets(ctx, tx, op)
	if err != nil {
		return op, err
	}
	writeback.ProjectTargets(&op)
	if op.Plan.Mirror != nil {
		op.Mirror, err = aiReadMirrorStep(ctx, tx, op)
		if err != nil {
			return op, err
		}
		writeback.ProjectMirror(&op)
	}
	return op, nil
}

func aiReadStackTargets(ctx context.Context, tx *sql.Tx, op writeback.Operation) ([]writeback.TargetOutcome, error) {
	rows, err := tx.QueryContext(ctx, `SELECT assetID,ordinal,status,code,attempts,generation,observed,verified,refreshed,noop,(completionKnown=1 AND senderActive=0) FROM ai_stack_write_targets WHERE userID=? AND installationID=? AND operationID=? ORDER BY ordinal`, op.Plan.Owner, op.Plan.Installation, op.ID)
	if err != nil {
		return nil, drafts.ErrStorage
	}
	defer rows.Close()
	var targets []writeback.TargetOutcome
	for rows.Next() {
		var t writeback.TargetOutcome
		var ordinal int
		var raw []byte
		if err = rows.Scan(&t.AssetID, &ordinal, &t.Status, &t.Code, &t.Attempts, &t.Generation, &raw, &t.Verified, &t.Refreshed, &t.Noop, &t.Settled); err != nil {
			return nil, drafts.ErrStorage
		}
		if ordinal != len(targets) || ordinal >= len(op.Plan.Manifest.Targets) || t.AssetID != op.Plan.Manifest.Targets[ordinal].AssetID || (len(raw) > 0 && json.Unmarshal(raw, &t.Observed) != nil) {
			return nil, drafts.ErrStorage
		}
		for _, field := range op.Plan.Manifest.Targets[ordinal].Fields {
			t.Fields = append(t.Fields, writeback.FieldOutcome{Field: field, Status: "pending"})
		}
		t.Events = []writeback.Event{}
		targets = append(targets, t)
	}
	if err = rows.Err(); err != nil || len(targets) != len(op.Plan.Manifest.Targets) {
		return nil, drafts.ErrStorage
	}
	if err = rows.Close(); err != nil {
		return nil, drafts.ErrStorage
	}
	for i := range targets {
		targets[i].Fields, err = aiReadStackFields(ctx, tx, op, i)
		if err != nil {
			return nil, err
		}
		targets[i].Events, err = aiReadStackEvents(ctx, tx, op, targets[i].AssetID)
		if err != nil {
			return nil, err
		}
	}
	return targets, nil
}

func aiReadStackEvents(ctx context.Context, tx *sql.Tx, op writeback.Operation, assetID string) ([]writeback.Event, error) {
	rows, err := tx.QueryContext(ctx, `SELECT code,at,attempt FROM ai_stack_write_events WHERE userID=? AND installationID=? AND operationID=? AND assetID=? ORDER BY sequence LIMIT 100`, op.Plan.Owner, op.Plan.Installation, op.ID, assetID)
	if err != nil {
		return nil, drafts.ErrStorage
	}
	defer rows.Close()
	events := []writeback.Event{}
	for rows.Next() {
		var e writeback.Event
		var at int64
		if rows.Scan(&e.Code, &at, &e.Attempt) != nil {
			return nil, drafts.ErrStorage
		}
		e.At = time.Unix(0, at).UTC().Format(time.RFC3339Nano)
		events = append(events, e)
	}
	return events, rows.Err()
}
