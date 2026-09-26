package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

func aiParseWriteMetadata(raw []byte, asset string, allowInvalidGPS bool) (writepreview.Metadata, string, error) {
	var result writepreview.Metadata
	if _, err := aiImageMetadataDigest(raw, asset, false); err != nil {
		return result, "", err
	}
	fields, err := aiImageUniqueObject(raw)
	if err != nil {
		return result, "", err
	}
	exif := fields["exifinfo"]
	if len(exif) == 0 {
		exif = json.RawMessage("null")
	}
	if string(exif) != "null" {
		if _, err = aiImageUniqueObject(exif); err != nil {
			return result, "", err
		}
		if json.Unmarshal(exif, &result.GPS) != nil || !aiDraftGPSValue(result.GPS.Latitude, 90) || !aiDraftGPSValue(result.GPS.Longitude, 180) {
			if !allowInvalidGPS {
				return result, "", drafts.ErrUnavailable
			}
			result.GPS, result.GPSUnavailable = writepreview.GPS{}, true
		}
	}
	var meta aiImageMetadata
	result.Description = aiPreviewText(aiDescriptionBaseline(exif))
	if json.Unmarshal(raw, &meta) != nil {
		return result, "", drafts.ErrUnavailable
	}
	identity, _ := json.Marshal([]string{"reviewed-image-v1", meta.ID, meta.OwnerID, meta.Type, meta.Checksum})
	sum := sha256.Sum256(identity)
	result.ImageIdentity = "v1:" + hex.EncodeToString(sum[:])
	stackID := ""
	if meta.Stack != nil {
		stackID = meta.Stack.ID
	}
	return result, stackID, nil
}
