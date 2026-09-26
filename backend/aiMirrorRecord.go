package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func (s *aiDraftStore) mirrorRecord(ctx context.Context, owner, asset string) (string, json.RawMessage, error) {
	authority, err := s.results.imageAuthority(ctx, owner, asset)
	if err != nil {
		return "", nil, drafts.ErrUnavailable
	}
	var id string
	var owned json.RawMessage
	err = s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := s.checkAuthority(ctx, tx, owner, asset, authority); err != nil {
			return err
		}
		var operation, value sql.NullString
		err := tx.QueryRowContext(ctx, `SELECT recordID,lastOperationID,value FROM ai_mirror_records WHERE userID=? AND installationID=? AND assetID=?`, owner, s.results.jobs.binding, asset).Scan(&id, &operation, &value)
		if errors.Is(err, sql.ErrNoRows) {
			id = uuid.NewString()
			_, err = tx.ExecContext(ctx, `INSERT INTO ai_mirror_records(userID,installationID,assetID,recordID) VALUES(?,?,?,?)`, owner, s.results.jobs.binding, asset, id)
			if err != nil {
				return drafts.ErrStorage
			}
			return nil
		}
		if err != nil {
			return drafts.ErrStorage
		}
		parsed, err := uuid.Parse(id)
		if err != nil || parsed.String() != id || parsed.Version() != 4 {
			return drafts.ErrStorage
		}
		if operation.Valid || value.Valid {
			var verified bool
			var payload, observed []byte
			var digest, previewID string
			if !operation.Valid || !value.Valid || tx.QueryRowContext(ctx, `SELECT m.verified,o.payload,o.digest,o.previewID,m.observed FROM ai_mirror_write_steps m JOIN ai_stack_write_operations o ON o.userID=m.userID AND o.installationID=m.installationID AND o.id=m.operationID WHERE m.userID=? AND m.installationID=? AND m.operationID=? AND m.assetID=? AND length(o.payload)<=1048576 AND length(m.observed)<=131072`, owner, s.results.jobs.binding, operation.String, asset).Scan(&verified, &payload, &digest, &previewID, &observed) != nil || !verified {
				return drafts.ErrStorage
			}
			plan, err := writeback.DecodePlan(payload, digest, owner, s.results.jobs.binding, previewID)
			var baseline writepreview.MirrorBaseline
			if err != nil || plan.TargetID != asset || plan.Mirror == nil || plan.Mirror.RecordID != id || json.Unmarshal(observed, &baseline) != nil || !baseline.Present || !writepreview.EqualMirrorValue(baseline.Value, plan.Mirror.Value) || !writepreview.EqualMirrorValue(json.RawMessage(value.String), plan.Mirror.Value) {
				return drafts.ErrStorage
			}
			owned = json.RawMessage(value.String)
		}
		return nil
	})
	if err != nil {
		return "", nil, err
	}
	return id, owned, nil
}
