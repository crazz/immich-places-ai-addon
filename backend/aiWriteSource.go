package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

type aiWriteAttempt struct {
	store     *aiWriteStore
	authority aiImageAuthority
	token     string
}

func (a *aiWriteAttempt) Read(ctx context.Context, op writeback.Operation) (writepreview.Metadata, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var result writepreview.Metadata
	s := a.store
	if s.images == nil {
		return result, drafts.ErrUnavailable
	}
	authority, err := s.readAuthority(ctx, op)
	if err != nil {
		return result, err
	}
	raw, _, err := s.images.fetch(ctx, authority.key, "/api/assets/"+op.Plan.TargetID, 1<<20)
	if err != nil {
		return result, err
	}
	defer clear(raw)
	if _, err = aiImageMetadataDigest(raw, op.Plan.TargetID, false); err != nil {
		return result, err
	}
	fields, err := aiImageUniqueObject(raw)
	if err != nil {
		return result, err
	}
	exif := fields["exifinfo"]
	if len(exif) > 0 && string(exif) != "null" {
		if _, err = aiImageUniqueObject(exif); err != nil {
			return result, err
		}
		if json.Unmarshal(exif, &result.GPS) != nil || !aiDraftGPSValue(result.GPS.Latitude, 90) || !aiDraftGPSValue(result.GPS.Longitude, 180) {
			return result, drafts.ErrUnavailable
		}
	}
	var meta aiImageMetadata
	if json.Unmarshal(raw, &meta) != nil {
		return result, drafts.ErrUnavailable
	}
	identity, _ := json.Marshal([]string{"reviewed-image-v1", meta.ID, meta.OwnerID, meta.Type, meta.Checksum})
	sum := sha256.Sum256(identity)
	result.ImageIdentity = "v1:" + hex.EncodeToString(sum[:])
	current, err := s.readAuthority(ctx, op)
	if err != nil || current != authority {
		return result, drafts.ErrUnavailable
	}
	a.authority = authority
	return result, nil
}
