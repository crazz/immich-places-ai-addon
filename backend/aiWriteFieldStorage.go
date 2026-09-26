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

func aiStoreWriteFields(ctx context.Context, tx *sql.Tx, op writeback.Operation, fields []writeback.FieldOutcome) error {
	if op.Plan.Version != "standard-preview-v2" {
		return nil
	}
	for _, field := range fields {
		var observed any = field.GPS
		if field.Field == "description" {
			observed = field.Description
		}
		raw, err := json.Marshal(observed)
		if err != nil {
			return drafts.ErrStorage
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO ai_standard_write_fields(userID,installationID,operationID,field,status,observed,wasVerified) VALUES(?,?,?,?,?,?,?) ON CONFLICT(userID,installationID,operationID,field) DO UPDATE SET status=excluded.status,observed=excluded.observed,wasVerified=max(ai_standard_write_fields.wasVerified,excluded.wasVerified)`, op.Plan.Owner, op.Plan.Installation, op.ID, field.Field, field.Status, string(raw), field.WasVerified); err != nil {
			return drafts.ErrStorage
		}
	}
	return nil
}

func aiReadWriteFields(ctx context.Context, tx *sql.Tx, op *writeback.Operation) error {
	if op.Plan.Version != "standard-preview-v2" {
		return nil
	}
	for _, field := range op.Plan.Fields {
		outcome := writeback.FieldOutcome{Field: field, Status: "pending"}
		var raw []byte
		err := tx.QueryRowContext(ctx, `SELECT status,observed,wasVerified FROM ai_standard_write_fields WHERE userID=? AND installationID=? AND operationID=? AND field=?`, op.Plan.Owner, op.Plan.Installation, op.ID, field).Scan(&outcome.Status, &raw, &outcome.WasVerified)
		if !errors.Is(err, sql.ErrNoRows) {
			if err != nil || len(raw) > 400000 {
				return drafts.ErrStorage
			}
			if field == "description" {
				err = json.Unmarshal(raw, &outcome.Description)
				if outcome.Description != nil && !writepreview.ValidTextObservation(*outcome.Description) {
					return drafts.ErrStorage
				}
			} else {
				err = json.Unmarshal(raw, &outcome.GPS)
			}
			if err != nil {
				return drafts.ErrStorage
			}
		}
		op.Fields = append(op.Fields, outcome)
	}
	return nil
}
