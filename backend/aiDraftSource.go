package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"

	"immich-places-backend/internal/ai/drafts"
)

func (s *aiDraftStore) source(ctx context.Context, owner, asset string, reader *aiImagePreparer) (drafts.Baseline, aiImageAuthority, error) {
	var baseline drafts.Baseline
	authority, err := s.results.imageAuthority(ctx, owner, asset)
	if err != nil || reader == nil {
		return baseline, authority, drafts.ErrUnavailable
	}
	data, _, err := reader.fetch(ctx, authority.key, "/api/assets/"+asset, 1<<20)
	if err != nil {
		return baseline, authority, drafts.ErrUnavailable
	}
	defer clear(data)
	digest, err := aiImageSourceDigest(data, asset)
	if err != nil {
		return baseline, authority, drafts.ErrUnavailable
	}
	fields, err := aiImageUniqueObject(data)
	if err != nil {
		return baseline, authority, drafts.ErrUnavailable
	}
	exif, exists := fields["exifinfo"]
	if !exists {
		exif = json.RawMessage("null")
	}
	if string(exif) != "null" {
		if _, err = aiImageUniqueObject(exif); err != nil {
			return baseline, authority, drafts.ErrUnavailable
		}
	}
	var gps struct {
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	}
	if json.Unmarshal(exif, &gps) != nil || !aiDraftGPSValue(gps.Latitude, 90) || !aiDraftGPSValue(gps.Longitude, 180) {
		return baseline, authority, drafts.ErrUnavailable
	}
	var meta aiImageMetadata
	if json.Unmarshal(data, &meta) != nil {
		return baseline, authority, drafts.ErrUnavailable
	}
	identity, _ := json.Marshal([]string{"reviewed-image-v1", meta.ID, meta.OwnerID, meta.Type, meta.Checksum})
	sum := sha256.Sum256(identity)
	baseline = drafts.Baseline{Status: "reviewed", Latitude: gps.Latitude, Longitude: gps.Longitude, ImageIdentity: "v1:" + hex.EncodeToString(sum[:]), SourceDigest: digest, AssetID: meta.ID, OwnerID: meta.OwnerID, Checksum: meta.Checksum, Type: meta.Type}
	if current, err := s.results.imageAuthority(ctx, owner, asset); err != nil || current != authority {
		return drafts.Baseline{}, authority, drafts.ErrUnavailable
	}
	return baseline, authority, nil
}
func aiDraftGPSValue(value *float64, limit float64) bool {
	return value == nil || (!math.IsNaN(*value) && !math.IsInf(*value, 0) && *value >= -limit && *value <= limit)
}
