package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/selection"
)

var errAISelectionDisabled = errors.New("AI is disabled")

type aiSelectionStore struct {
	enabled   bool
	binding   string
	db        *Database
	now       func() time.Time
	maxAssets int
	ttl       time.Duration
}

func newAISelectionStore(db *Database) *aiSelectionStore {
	return &aiSelectionStore{db: db, now: time.Now, maxAssets: 500, ttl: 15 * time.Minute}
}

func (s *aiSelectionStore) preview(ctx context.Context, input selection.Input, owner string) (selection.Manifest, error) {
	if !s.enabled {
		return selection.Manifest{}, errAISelectionDisabled
	}
	request, err := selection.Normalize(input, s.maxAssets)
	if err != nil {
		return selection.Manifest{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.cleanup(ctx); err != nil {
		return selection.Manifest{}, err
	}
	tx, err := s.db.db.BeginTx(ctx, nil)
	if err != nil {
		return selection.Manifest{}, err
	}
	defer tx.Rollback()
	if err := s.checkBinding(ctx, tx); err != nil {
		return selection.Manifest{}, err
	}
	if err := validateAISelectionScope(ctx, tx, owner, *request.Scope); err != nil {
		return selection.Manifest{}, err
	}
	now := s.now().UTC()
	id := uuid.NewString()
	manifest := selection.Manifest{SnapshotID: &id, Mode: request.Mode, Scope: *request.Scope, PolicyVersion: selection.PolicyVersion,
		AssetIDs: []string{}, Exclusions: []selection.Exclusion{}, RequestedCount: request.RequestedCount,
		UniqueCount: len(request.AssetIDs), DuplicateCount: request.DuplicateCount, CreatedAt: now, ExpiresAt: now.Add(s.ttl)}
	facts := make([]string, 0, len(request.AssetIDs))
	factsBytes := 0
	for _, assetID := range request.AssetIDs {
		candidate, err := resolveAISelectionCandidate(ctx, tx, owner, assetID, manifest.Scope)
		if err != nil {
			return selection.Manifest{}, err
		}
		reason := selection.ExclusionReason(candidate.Candidate)
		if reason != "" {
			manifest.Exclusions = append(manifest.Exclusions, selection.Exclusion{AssetID: assetID, Reason: reason})
			continue
		}
		fact, err := aiSelectionFacts(candidate)
		if err != nil {
			return selection.Manifest{}, err
		}
		factsBytes += len(fact)
		if factsBytes > selection.MaxBytes {
			return selection.Manifest{}, selection.ErrLimit
		}
		facts = append(facts, fact)
		manifest.AssetIDs = append(manifest.AssetIDs, assetID)
	}
	manifest.EligibleCount = len(manifest.AssetIDs)
	manifest.ExcludedCount = len(manifest.Exclusions)
	if manifest.EligibleCount == 0 {
		manifest.SnapshotID = nil
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		return selection.Manifest{}, err
	}
	if len(data)+factsBytes > selection.MaxBytes {
		return selection.Manifest{}, selection.ErrLimit
	}
	if manifest.EligibleCount == 0 {
		return manifest, nil
	}
	if err := aiSelectionQuota(ctx, tx, owner, now); err != nil {
		return selection.Manifest{}, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO ai_selection_snapshots (userID,id,expiresAt,manifest,digest) VALUES (?,?,?,?,?)", owner, id, manifest.ExpiresAt.UnixNano(), string(data), selection.Digest(s.binding, owner, string(data), facts)); err != nil {
		return selection.Manifest{}, err
	}
	for position, assetID := range manifest.AssetIDs {
		if _, err = tx.ExecContext(ctx, "INSERT INTO ai_selection_items (userID,snapshotID,position,assetID,facts) VALUES (?,?,?,?,?)", owner, id, position, assetID, facts[position]); err != nil {
			return selection.Manifest{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return selection.Manifest{}, err
	}
	return manifest, nil
}
