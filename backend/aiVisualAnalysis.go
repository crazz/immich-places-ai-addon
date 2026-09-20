package main

import (
	"context"
	"errors"
	"time"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/images"
	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

type aiVisualAnalyzer struct {
	db         *Database
	images     *aiImagePreparer
	dispatcher *providers.Dispatcher
	runner     *analysis.Runner
}

func newAIVisualAnalyzer(db *Database, images *aiImagePreparer, dispatcher *providers.Dispatcher) (*aiVisualAnalyzer, error) {
	if db == nil || images == nil || dispatcher == nil {
		return nil, analysis.ErrInvalidRequest
	}
	a := &aiVisualAnalyzer{db: db, images: images, dispatcher: dispatcher}
	runner, err := analysis.New(analysis.Protocol{Encode: providerhttp.EncodeVisual, Parse: providerhttp.ParseVisual}, func(ctx context.Context, input analysis.DispatchRequest) (analysis.DispatchReply, error) {
		reply, err := a.dispatcher.Dispatch(ctx, providers.DispatchRequest{OwnerID: input.OwnerID, ProfileID: input.ProfileID, Revision: input.Revision, Body: input.Body, Authorize: input.Authorize})
		if err != nil {
			return analysis.DispatchReply{}, providerhttp.ClassifyVisualFailure(err)
		}
		return analysis.DispatchReply{Body: reply.Body}, nil
	})
	if err != nil {
		return nil, err
	}
	a.runner = runner
	return a, nil
}
func (a *aiVisualAnalyzer) analyze(ctx context.Context, req analysis.Request) (result *analysis.Result, err error) {
	return a.analyzeWith(ctx, req, a.runner.RunVisual)
}

func (a *aiVisualAnalyzer) analyzeWith(ctx context.Context, req analysis.Request, run func(context.Context, analysis.Request) (*analysis.Result, error)) (result *analysis.Result, err error) {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	defer func() {
		if ctx.Err() != nil {
			result, err = nil, ctx.Err()
		}
	}()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if req.Guard.Authorize == nil || req.Guard.Reserve == nil {
		return nil, analysis.ErrInvalidRequest
	}
	version, err := a.loadProfile(ctx, req)
	if err != nil {
		return nil, analysis.ErrDenied
	}
	req.Model = version.model
	authority, err := a.images.authorize(ctx, req.Owner, req.Installation, req.Asset)
	if err != nil {
		return nil, analysis.ErrDenied
	}
	imageInfo, valid := req.Image.Info()
	if !valid {
		return nil, analysis.ErrInvalidRequest
	}
	callerAuthorize := req.Guard.Authorize
	req.Guard.Authorize = func(ctx context.Context) error {
		if err := callerAuthorize(ctx); err != nil {
			return analysis.ErrDenied
		}
		current, err := a.loadProfile(ctx, req)
		if err != nil || current != version {
			return analysis.ErrDenied
		}
		local, err := a.images.authorize(ctx, req.Owner, req.Installation, req.Asset)
		if err != nil || local != authority {
			return analysis.ErrDenied
		}
		source, err := a.images.sourceDigest(ctx, local.key, req.Asset)
		if err != nil {
			if errors.Is(err, images.ErrLimit) {
				return analysis.ErrLimit
			}
			if errors.Is(err, errAIImageUpstream) {
				return analysis.ErrUpstream
			}
			return analysis.ErrDenied
		}
		if source != imageInfo.Binding.SourceDigest {
			return analysis.ErrDenied
		}
		local, err = a.images.authorize(ctx, req.Owner, req.Installation, req.Asset)
		if err != nil || local != authority {
			return analysis.ErrDenied
		}
		info, valid := req.Image.Info()
		if !valid || info != imageInfo {
			return analysis.ErrDenied
		}
		return nil
	}
	return run(ctx, req)
}
