package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/selection"
)

type aiImageAuthority struct{ key, facts string }

func (s *aiImagePreparer) authorize(ctx context.Context, owner, installation, asset string) (aiImageAuthority, error) {
	if !s.selections.enabled || owner == "" || len(owner) > 128 || installation != s.selections.binding {
		return aiImageAuthority{}, errAIImageDenied
	}
	for _, id := range []string{installation, asset} {
		parsed, err := uuid.Parse(id)
		if err != nil || parsed.String() != id {
			return aiImageAuthority{}, errAIImageDenied
		}
	}
	var current string
	if err := s.db.db.QueryRowContext(ctx, "SELECT id FROM ai_installation_identity WHERE singleton=1").Scan(&current); err != nil || current != installation {
		return aiImageAuthority{}, errAIImageDenied
	}
	user, err := s.db.getUserByID(ctx, owner)
	if err != nil || user == nil || user.ImmichAPIKey == nil || strings.TrimSpace(*user.ImmichAPIKey) == "" {
		return aiImageAuthority{}, errAIImageDenied
	}
	row, err := s.db.getAssetByID(ctx, owner, asset)
	if err != nil || row == nil {
		return aiImageAuthority{}, errAIImageDenied
	}
	candidate := selection.Candidate{Available: true, Type: row.Type, Hidden: row.IsHidden, StackChild: row.StackPrimaryAssetID != nil && *row.StackPrimaryAssetID != "" && *row.StackPrimaryAssetID != asset, InScope: true}
	if selection.ExclusionReason(candidate) != "" {
		return aiImageAuthority{}, errAIImageDenied
	}
	facts, err := json.Marshal(struct {
		Type                      string
		Hidden                    bool
		StackID, Primary, Library *string
		StackCount                *int
	}{row.Type, row.IsHidden, row.StackID, row.StackPrimaryAssetID, row.LibraryID, row.StackAssetCount})
	if err != nil {
		return aiImageAuthority{}, errAIImageDenied
	}
	digest := sha256.Sum256(facts)
	return aiImageAuthority{key: *user.ImmichAPIKey, facts: hex.EncodeToString(digest[:])}, nil
}
