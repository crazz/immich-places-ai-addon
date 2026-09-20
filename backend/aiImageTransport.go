package main

import (
	"context"
	"io"
	"net/http"

	"immich-places-backend/internal/ai/images"
)

func (s *aiImagePreparer) fetch(ctx context.Context, key, path string, limit int) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.endpoint+path, nil)
	if err != nil {
		return nil, "", errAIImageUpstream
	}
	req.Header.Set("x-api-key", key)
	req.Header.Set("Accept-Encoding", "identity")
	response, err := s.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, "", ctx.Err()
		}
		return nil, "", errAIImageUpstream
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, "", aiImageHTTPError{Status: response.StatusCode}
	}
	if encoding := response.Header.Get("Content-Encoding"); encoding != "" && encoding != "identity" {
		return nil, "", errAIImageUpstream
	}
	if limit <= 0 || response.ContentLength > int64(limit) {
		return nil, "", images.ErrLimit
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, int64(limit)+1))
	if err != nil {
		clear(data)
		if ctx.Err() != nil {
			return nil, "", ctx.Err()
		}
		return nil, "", errAIImageUpstream
	}
	if len(data) > limit {
		clear(data)
		return nil, "", images.ErrLimit
	}
	return data, response.Header.Get("Content-Type"), nil
}

type aiImageHTTPError struct{ Status int }

func (aiImageHTTPError) Error() string { return "image source unavailable" }
func (aiImageHTTPError) Unwrap() error { return errAIImageUpstream }
