package main

import (
	"context"
	"database/sql"
	"immich-places-backend/internal/ai/selection"
)

func resolveAIExplicitSelection(ctx context.Context, tx *sql.Tx, owner string, request selection.Request, manifest *selection.Manifest) ([]string, int, error) {
	facts := make([]string, 0, len(request.AssetIDs))
	factsBytes := 0
	for _, assetID := range request.AssetIDs {
		candidate, err := resolveAISelectionCandidate(ctx, tx, owner, assetID, manifest.Scope)
		if err != nil {
			return nil, 0, err
		}
		reason := selection.ExclusionReason(candidate.Candidate)
		if reason != "" {
			manifest.Exclusions = append(manifest.Exclusions, selection.Exclusion{AssetID: assetID, Reason: reason})
			continue
		}
		fact, err := aiSelectionFacts(candidate)
		if err != nil {
			return nil, 0, err
		}
		factsBytes += len(fact)
		if factsBytes > selection.MaxBytes {
			return nil, 0, selection.ErrLimit
		}
		facts = append(facts, fact)
		manifest.AssetIDs = append(manifest.AssetIDs, assetID)
	}
	manifest.EligibleCount = len(manifest.AssetIDs)
	manifest.ExcludedCount = len(manifest.Exclusions)
	return facts, factsBytes, nil
}
