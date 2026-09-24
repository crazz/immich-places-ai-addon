package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"immich-places-backend/internal/ai/drafts"
)

func (s *aiDraftStore) checkAuthority(ctx context.Context, tx *sql.Tx, owner, asset string, expected aiImageAuthority) error {
	var user UserRow
	if err := tx.QueryRowContext(ctx, `SELECT immichAPIKey FROM users WHERE ID=?`, owner).Scan(&user.ImmichAPIKey); err != nil || user.ImmichAPIKey == nil {
		return drafts.ErrUnavailable
	}
	if s.results.jobs.db.decryptUserSecrets(&user) != nil || user.ImmichAPIKey == nil || *user.ImmichAPIKey != expected.key {
		return drafts.ErrUnavailable
	}
	var facts struct {
		Type                    string
		Hidden                  bool
		Library, Stack, Primary *string
	}
	err := tx.QueryRowContext(ctx, `SELECT type,isHidden,libraryID,stackID,stackPrimaryAssetID FROM assets WHERE userID=? AND immichID=?`+hiddenLibraryFilter, owner, asset).Scan(&facts.Type, &facts.Hidden, &facts.Library, &facts.Stack, &facts.Primary)
	if err != nil || facts.Type != "IMAGE" || facts.Hidden {
		return drafts.ErrUnavailable
	}
	raw, err := json.Marshal(facts)
	if err != nil || string(raw) != expected.facts {
		return drafts.ErrUnavailable
	}
	return nil
}
