package main

import (
	"bytes"
	"encoding/json"
)

func aiContextExif(data []byte) (*ImmichExifInfo, error) {
	object, err := aiImageUniqueObject(data)
	if err != nil {
		return nil, errAIImageUpstream
	}
	raw := object["exifinfo"]
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	if _, err := aiImageUniqueObject(raw); err != nil {
		return nil, errAIImageUpstream
	}
	var exif ImmichExifInfo
	if json.Unmarshal(raw, &exif) != nil {
		return nil, errAIImageUpstream
	}
	return &exif, nil
}
