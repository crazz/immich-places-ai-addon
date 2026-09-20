package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"unicode/utf8"

	"immich-places-backend/internal/ai/images"
)

// Only the fixed metadata-search endpoint accepts a read-only POST.
func (s *aiContextPreparer) read(ctx context.Context, key, path string, search []byte) ([]byte, bool, error) {
	method := http.MethodGet
	if search != nil {
		if path != "/api/search/metadata" {
			return nil, false, errAIImageDenied
		}
		method = http.MethodPost
	}
	req, err := http.NewRequestWithContext(ctx, method, s.images.endpoint+path, bytes.NewReader(search))
	if err != nil {
		return nil, false, errAIImageUpstream
	}
	req.Header.Set("x-api-key", key)
	req.Header.Set("Accept-Encoding", "identity")
	if search != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := s.images.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, false, ctx.Err()
		}
		return nil, false, errAIImageUpstream
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusNotFound {
		return nil, false, nil
	}
	if response.StatusCode != http.StatusOK {
		return nil, false, errAIImageUpstream
	}
	if encoding := response.Header.Get("Content-Encoding"); encoding != "" && encoding != "identity" {
		return nil, false, errAIImageUpstream
	}
	const limit = 1 << 20
	if response.ContentLength > limit {
		return nil, false, images.ErrLimit
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		clear(data)
		if ctx.Err() != nil {
			return nil, false, ctx.Err()
		}
		return nil, false, errAIImageUpstream
	}
	if len(data) > limit {
		clear(data)
		return nil, false, images.ErrLimit
	}
	if !utf8.Valid(data) {
		clear(data)
		return nil, false, errAIImageUpstream
	}
	return data, true, nil
}
