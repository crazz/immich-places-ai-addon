package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func (s *aiWriteStore) readAuthority(ctx context.Context, op writeback.Operation) (aiImageAuthority, error) {
	var authority aiImageAuthority
	err := s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		authority, err = s.authority(ctx, tx, op)
		return err
	})
	return authority, err
}

func (s *aiWriteStore) authority(ctx context.Context, tx *sql.Tx, op writeback.Operation) (aiImageAuthority, error) {
	var a aiImageAuthority
	key, err := s.credential(ctx, tx, op.Plan.Owner)
	if err != nil {
		return a, err
	}
	var hash string
	if tx.QueryRowContext(ctx, aiWriteStatement(op.Plan, `SELECT credentialHash FROM ai_write_operations WHERE userID=? AND installationID=? AND id=?`), op.Plan.Owner, op.Plan.Installation, op.ID).Scan(&hash) != nil || hash != aiWriteCredentialHash(key) {
		return a, drafts.ErrUnavailable
	}
	var facts struct {
		Type                    string
		Hidden                  bool
		Library, Stack, Primary *string
	}
	err = tx.QueryRowContext(ctx, `SELECT type,isHidden,libraryID,stackID,stackPrimaryAssetID FROM assets WHERE userID=? AND immichID=?`+hiddenLibraryFilter, op.Plan.Owner, op.Plan.TargetID).Scan(&facts.Type, &facts.Hidden, &facts.Library, &facts.Stack, &facts.Primary)
	if errors.Is(err, sql.ErrNoRows) && op.Status != "queued" {
		var exists int
		if tx.QueryRowContext(ctx, `SELECT count(*) FROM assets WHERE userID=? AND immichID=?`, op.Plan.Owner, op.Plan.TargetID).Scan(&exists) != nil || exists != 0 {
			return a, drafts.ErrUnavailable
		}
		return aiImageAuthority{key: key, facts: "missing"}, nil
	}
	if err != nil || facts.Type != "IMAGE" || facts.Hidden {
		return a, drafts.ErrUnavailable
	}
	raw, err := json.Marshal(facts)
	if err != nil {
		return a, drafts.ErrStorage
	}
	return aiImageAuthority{key: key, facts: string(raw)}, nil
}
