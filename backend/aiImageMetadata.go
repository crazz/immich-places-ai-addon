package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

type aiImageMetadata struct {
	ID         string `json:"id"`
	OwnerID    string `json:"ownerId"`
	Type       string `json:"type"`
	Visibility string `json:"visibility"`
	IsTrashed  *bool  `json:"isTrashed"`
	Checksum   string `json:"checksum"`
	UpdatedAt  string `json:"updatedAt"`
	Stack      *struct {
		ID             string `json:"id"`
		PrimaryAssetID string `json:"primaryAssetId"`
		AssetCount     int    `json:"assetCount"`
	} `json:"stack"`
}

func aiImageSourceDigest(data []byte, asset string) (string, error) {
	fields, err := aiImageUniqueObject(data)
	if err != nil {
		return "", err
	}
	if stack := fields["stack"]; len(stack) != 0 && string(stack) != "null" {
		if _, err := aiImageUniqueObject(stack); err != nil {
			return "", err
		}
	}
	var meta aiImageMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return "", errAIImageUpstream
	}
	if meta.ID != asset || meta.Type != "IMAGE" || (meta.Visibility != "timeline" && meta.Visibility != "archive") || meta.IsTrashed == nil || *meta.IsTrashed {
		return "", errAIImageDenied
	}
	if _, err := uuid.Parse(meta.OwnerID); err != nil {
		return "", errAIImageUpstream
	}
	checksum, err := base64.StdEncoding.DecodeString(meta.Checksum)
	if err != nil || len(checksum) != 20 {
		return "", errAIImageUpstream
	}
	if _, err := time.Parse(time.RFC3339Nano, meta.UpdatedAt); err != nil {
		return "", errAIImageUpstream
	}
	if meta.Stack != nil {
		if _, err := uuid.Parse(meta.Stack.ID); err != nil || meta.Stack.PrimaryAssetID != asset || meta.Stack.AssetCount < 1 {
			return "", errAIImageDenied
		}
	}
	canonical, err := json.Marshal(meta)
	if err != nil {
		return "", errAIImageUpstream
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}

func aiImageUniqueObject(data []byte) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, errAIImageUpstream
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err = decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return nil, errAIImageUpstream
		}
		key = strings.ToLower(key)
		if _, exists := fields[key]; exists {
			return nil, errAIImageUpstream
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, errAIImageUpstream
		}
		fields[key] = value
	}
	if token, err = decoder.Token(); err != nil || token != json.Delim('}') {
		return nil, errAIImageUpstream
	}
	if _, err = decoder.Token(); err != io.EOF {
		return nil, errAIImageUpstream
	}
	return fields, nil
}

func (s *aiImagePreparer) sourceDigest(ctx context.Context, key, asset string) (string, error) {
	data, _, err := s.fetch(ctx, key, "/api/assets/"+asset, 1<<20)
	if err != nil {
		return "", err
	}
	defer clear(data)
	return aiImageSourceDigest(data, asset)
}
