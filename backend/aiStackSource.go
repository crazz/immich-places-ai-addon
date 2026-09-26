package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

type aiStackSource struct {
	metadata           writepreview.Metadata
	authority          aiImageAuthority
	stackID, primaryID string
	memberCount        int
}

func (s *aiDraftStore) stackSource(ctx context.Context, owner, asset string, reader *aiImagePreparer) (aiStackSource, error) {
	var result aiStackSource
	if reader == nil {
		return result, drafts.ErrUnavailable
	}
	authority, err := s.results.imageAuthority(ctx, owner, asset)
	if err != nil {
		return result, drafts.ErrUnavailable
	}
	raw, _, err := reader.fetch(ctx, authority.key, "/api/assets/"+asset, 1<<20)
	if err != nil {
		return result, drafts.ErrUnavailable
	}
	defer clear(raw)
	if _, err = aiImageMetadataDigest(raw, asset, false); err != nil {
		return result, drafts.ErrUnavailable
	}
	fields, err := aiImageUniqueObject(raw)
	if err != nil {
		return result, drafts.ErrUnavailable
	}
	exif := fields["exifinfo"]
	if len(exif) == 0 {
		exif = json.RawMessage("null")
	}
	if string(exif) != "null" {
		if _, err = aiImageUniqueObject(exif); err != nil {
			return result, drafts.ErrUnavailable
		}
	}
	if json.Unmarshal(exif, &result.metadata.GPS) != nil || !aiDraftGPSValue(result.metadata.GPS.Latitude, 90) || !aiDraftGPSValue(result.metadata.GPS.Longitude, 180) {
		return result, drafts.ErrUnavailable
	}
	var meta aiImageMetadata
	if json.Unmarshal(raw, &meta) != nil || meta.Stack == nil {
		return result, drafts.ErrUnavailable
	}
	identity, _ := json.Marshal([]string{"reviewed-image-v1", meta.ID, meta.OwnerID, meta.Type, meta.Checksum})
	sum := sha256.Sum256(identity)
	result.metadata.ImageIdentity = "v1:" + hex.EncodeToString(sum[:])
	result.metadata.Description = aiPreviewText(aiDescriptionBaseline(exif))
	result.stackID, result.primaryID, result.memberCount = meta.Stack.ID, meta.Stack.PrimaryAssetID, meta.Stack.AssetCount
	if current, err := s.results.imageAuthority(ctx, owner, asset); err != nil || current != authority {
		return aiStackSource{}, drafts.ErrUnavailable
	}
	result.authority = authority
	return result, nil
}

func (s *aiDraftStore) stackMembers(ctx context.Context, owner string, source aiStackSource, reader *aiImagePreparer) ([]string, []string, error) {
	raw, _, err := reader.fetch(ctx, source.authority.key, "/api/stacks/"+source.stackID, 1<<20)
	if err != nil {
		return nil, nil, drafts.ErrUnavailable
	}
	defer clear(raw)
	if _, err = aiImageUniqueObject(raw); err != nil {
		return nil, nil, drafts.ErrUnavailable
	}
	var value struct {
		ID        string            `json:"id"`
		PrimaryID string            `json:"primaryAssetId"`
		Assets    []json.RawMessage `json:"assets"`
	}
	if json.Unmarshal(raw, &value) != nil || value.ID != source.stackID || value.PrimaryID != source.primaryID || len(value.Assets) != source.memberCount {
		return nil, nil, drafts.ErrUnavailable
	}
	ids, candidates := make([]string, 0, len(value.Assets)), make([]string, 0, len(value.Assets))
	for _, item := range value.Assets {
		if _, err := aiImageUniqueObject(item); err != nil {
			return nil, nil, drafts.ErrUnavailable
		}
		var meta aiImageMetadata
		if json.Unmarshal(item, &meta) != nil {
			return nil, nil, drafts.ErrUnavailable
		}
		if parsed, err := uuid.Parse(meta.ID); err != nil || parsed.String() != meta.ID || slices.Contains(ids, meta.ID) {
			return nil, nil, drafts.ErrUnavailable
		}
		ids = append(ids, meta.ID)
		if meta.Type != "IMAGE" || meta.IsTrashed == nil || *meta.IsTrashed || (meta.Visibility != "timeline" && meta.Visibility != "archive") {
			continue
		}
		if _, err := s.results.imageAuthority(ctx, owner, meta.ID); err == nil {
			candidates = append(candidates, meta.ID)
		}
	}
	slices.Sort(ids)
	slices.Sort(candidates)
	return ids, candidates, nil
}
