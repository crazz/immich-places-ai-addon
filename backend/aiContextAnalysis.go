package main

import (
	"context"
	"errors"
	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/images"
	"slices"
)

func (a *aiVisualAnalyzer) analyzeContext(ctx context.Context, req analysis.Request, preparer *aiContextPreparer, contextReq aiContextRequest) (*analysis.Result, error) {
	return a.analyzeContextWith(ctx, req, preparer, contextReq, a.runner.RunContext)
}

func (a *aiVisualAnalyzer) analyzeContextWith(ctx context.Context, req analysis.Request, preparer *aiContextPreparer, contextReq aiContextRequest, run func(context.Context, analysis.Request) (*analysis.Result, error)) (*analysis.Result, error) {
	if preparer == nil || preparer.images != a.images || req.Context == nil || req.Guard.Authorize == nil || contextReq.Authorize == nil || req.Context.Info().Binding != contextReq.Binding {
		return nil, analysis.ErrInvalidRequest
	}
	contextReq.Consent.Classes = slices.Clone(contextReq.Consent.Classes)
	callerAuthorize := req.Guard.Authorize
	var contextChanged bool
	req.Guard.Authorize = func(ctx context.Context) error {
		if err := callerAuthorize(ctx); err != nil {
			return analysis.ErrDenied
		}
		if err := preparer.check(ctx, contextReq, req.Context); err != nil {
			contextChanged = errors.Is(err, errAIContextChanged)
			if errors.Is(err, images.ErrLimit) {
				return analysis.ErrLimit
			}
			if errors.Is(err, errAIImageUpstream) {
				return analysis.ErrUpstream
			}
			return analysis.ErrDenied
		}
		return nil
	}
	result, err := a.analyzeWith(ctx, req, run)
	if contextChanged {
		return nil, errAIContextChanged
	}
	return result, err
}
