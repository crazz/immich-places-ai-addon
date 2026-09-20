package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"immich-places-backend/internal/ai/images"
)

var errAIImageDenied = errors.New("image access unavailable")
var errAIImageUpstream = errors.New("image source unavailable")

type aiImagePreparer struct {
	db         *Database
	selections *aiSelectionStore
	endpoint   string
	client     *http.Client
	limits     images.Limits
	timeout    time.Duration
}

func newAIImagePreparer(db *Database, store *aiSelectionStore, endpoint string) (*aiImagePreparer, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Hostname() == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return nil, images.ErrInvalid
	}
	transport := &http.Transport{
		DialContext:         (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second, ResponseHeaderTimeout: 10 * time.Second,
		DisableCompression: true, DisableKeepAlives: true, MaxResponseHeaderBytes: 64 << 10,
	}
	client := &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return &aiImagePreparer{db: db, selections: store, endpoint: strings.TrimRight(endpoint, "/"), client: client, limits: images.DefaultLimits(), timeout: 30 * time.Second}, nil
}
func (s *aiImagePreparer) prepare(ctx context.Context, owner, installation, asset string) (result *images.Prepared, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.timeout <= 0 || s.timeout > 30*time.Second {
		return nil, images.ErrInvalid
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	defer func() {
		if ctx.Err() != nil {
			result.Release()
			result, err = nil, ctx.Err()
		}
	}()
	authority, err := s.authorize(ctx, owner, installation, asset)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}
	source, err := s.sourceDigest(ctx, authority.key, asset)
	if err != nil {
		return nil, err
	}
	if current, err := s.authorize(ctx, owner, installation, asset); err != nil || current != authority {
		return nil, errAIImageDenied
	}
	data, mime, err := s.fetch(ctx, authority.key, "/api/assets/"+asset+"/thumbnail?size=preview", s.limits.MaxSourceBytes)
	if err != nil {
		return nil, err
	}
	defer clear(data)
	prepared, err := images.PrepareRaster(ctx, data, mime, images.Binding{Owner: owner, Installation: installation, Asset: asset, SourceDigest: source}, s.limits)
	if err != nil {
		return nil, err
	}
	if current, err := s.authorize(ctx, owner, installation, asset); err != nil || current != authority {
		prepared.Release()
		return nil, errAIImageDenied
	}
	currentSource, err := s.sourceDigest(ctx, authority.key, asset)
	if err != nil {
		prepared.Release()
		return nil, err
	}
	if currentSource != source {
		prepared.Release()
		return nil, errAIImageDenied
	}
	if current, err := s.authorize(ctx, owner, installation, asset); err != nil || current != authority {
		prepared.Release()
		return nil, errAIImageDenied
	}
	if err := ctx.Err(); err != nil {
		prepared.Release()
		return nil, err
	}
	return prepared, nil
}
