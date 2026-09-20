package main

import (
	"context"
	"errors"
	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/images"
	"slices"
)

func (a *aiVisualAnalyzer) analyzeContext(ctx context.Context, req analysis.Request, preparer *aiContextPreparer, contextReq aiContextRequest) (*analysis.Result, error) {
	if preparer == nil || preparer.images != a.images || req.Context == nil || req.Guard.Authorize == nil || contextReq.Authorize == nil || req.Context.Info().Binding != contextReq.Binding {
		return nil, analysis.ErrInvalidRequest
	}
	contextReq.Consent.Classes = slices.Clone(contextReq.Consent.Classes)
	callerAuthorize := req.Guard.Authorize
	req.Guard.Authorize = func(ctx context.Context) error {
		if err := callerAuthorize(ctx); err != nil {
			return analysis.ErrDenied
		}
		if err := preparer.check(ctx, contextReq, req.Context); err != nil {
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
	return a.analyzeWith(ctx, req, a.runner.RunContext)
}
