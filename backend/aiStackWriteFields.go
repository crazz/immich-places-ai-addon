package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func aiStoreStackFields(ctx context.Context, tx *sql.Tx, op writeback.Operation, fields []writeback.FieldOutcome) error {
	for _, field := range fields {
		var observed any = field.GPS
		if field.Field == "description" {
			observed = field.Description
		}
		raw, err := json.Marshal(observed)
		if err != nil {
			return drafts.ErrStorage
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO ai_stack_write_fields(userID,installationID,operationID,assetID,field,status,observed,wasVerified) VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(userID,installationID,operationID,assetID,field) DO UPDATE SET status=excluded.status,observed=excluded.observed,wasVerified=max(ai_stack_write_fields.wasVerified,excluded.wasVerified)`, op.Plan.Owner, op.Plan.Installation, op.ID, op.Plan.TargetID, field.Field, field.Status, string(raw), field.WasVerified); err != nil {
			return drafts.ErrStorage
		}
	}
	return nil
}

func aiReadStackFields(ctx context.Context, tx *sql.Tx, op writeback.Operation, ordinal int) ([]writeback.FieldOutcome, error) {
	target := op.Plan.Manifest.Targets[ordinal]
	fields := make([]writeback.FieldOutcome, 0, len(target.Fields))
	for _, field := range target.Fields {
		outcome := writeback.FieldOutcome{Field: field, Status: "pending"}
		var raw []byte
		err := tx.QueryRowContext(ctx, `SELECT status,observed,wasVerified FROM ai_stack_write_fields WHERE userID=? AND installationID=? AND operationID=? AND assetID=? AND field=?`, op.Plan.Owner, op.Plan.Installation, op.ID, target.AssetID, field).Scan(&outcome.Status, &raw, &outcome.WasVerified)
		if !errors.Is(err, sql.ErrNoRows) {
			if err != nil || len(raw) > 400000 {
				return nil, drafts.ErrStorage
			}
			if field == "description" {
				err = json.Unmarshal(raw, &outcome.Description)
				if outcome.Description != nil && !writepreview.ValidTextObservation(*outcome.Description) {
					return nil, drafts.ErrStorage
				}
			} else {
				err = json.Unmarshal(raw, &outcome.GPS)
			}
			if err != nil {
				return nil, drafts.ErrStorage
			}
		}
		fields = append(fields, outcome)
	}
	return fields, nil
}
