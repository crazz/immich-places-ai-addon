package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"immich-places-backend/internal/ai/selection"
)

var errAISelectionStale = errors.New("selection is stale")

type aiSelectionItem struct{ id, facts string }

func (s *aiSelectionStore) load(ctx context.Context, owner, id string) (selection.Manifest, error) {
	if !s.enabled {
		return selection.Manifest{}, errAISelectionDisabled
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tx, err := s.db.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return selection.Manifest{}, err
	}
	defer tx.Rollback()
	manifest, err := s.loadInTx(ctx, tx, owner, id)
	if err != nil {
		return selection.Manifest{}, err
	}

	if err = tx.Commit(); err != nil {
		return selection.Manifest{}, err
	}
	return manifest, nil
}

func (s *aiSelectionStore) loadInTx(ctx context.Context, tx *sql.Tx, owner, id string) (selection.Manifest, error) {
	var err error
	if err := s.checkBinding(ctx, tx); err != nil {
		return selection.Manifest{}, err
	}
	var data, digest string
	if err = tx.QueryRowContext(ctx, "SELECT manifest,digest FROM ai_selection_snapshots WHERE userID=? AND id=?", owner, id).Scan(&data, &digest); err != nil {
		return selection.Manifest{}, err
	}
	var manifest selection.Manifest
	if err = json.Unmarshal([]byte(data), &manifest); err != nil {
		return selection.Manifest{}, err
	}
	if !manifest.Current(s.now()) {
		return selection.Manifest{}, errAISelectionStale
	}
	items, err := loadAISelectionItems(ctx, tx, owner, id)
	if err != nil {
		return selection.Manifest{}, err
	}
	if len(items) != len(manifest.AssetIDs) {
		return selection.Manifest{}, errAISelectionStale
	}
	factsForDigest := make([]string, len(items))
	for i, item := range items {
		factsForDigest[i] = item.facts
	}
	if selection.Digest(s.binding, owner, data, factsForDigest) != digest {
		return selection.Manifest{}, errAISelectionStale
	}
	for i, item := range items {
		candidate, err := resolveAISelectionCandidate(ctx, tx, owner, item.id, manifest.Scope)
		if err != nil {
			return selection.Manifest{}, err
		}
		facts, err := aiSelectionFacts(candidate)
		if err != nil {
			return selection.Manifest{}, err
		}
		if item.id != manifest.AssetIDs[i] || selection.ExclusionReason(candidate.Candidate) != "" || facts != item.facts {
			return selection.Manifest{}, errAISelectionStale
		}
	}
	return manifest, nil
}

func loadAISelectionItems(ctx context.Context, tx *sql.Tx, owner, id string) ([]aiSelectionItem, error) {
	rows, err := tx.QueryContext(ctx, "SELECT assetID,facts FROM ai_selection_items WHERE userID=? AND snapshotID=? ORDER BY position", owner, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []aiSelectionItem{}
	for rows.Next() {
		var item aiSelectionItem
		if err := rows.Scan(&item.id, &item.facts); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
