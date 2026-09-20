package main

import (
	"context"
	"database/sql"
	"errors"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func (p *aiProductionJobs) executor(a *aiVisualAnalyzer) jobs.Execute {
	return func(ctx context.Context, lease jobs.Lease, guard jobs.Guard) (jobs.Completion, error) {
		var req jobs.Admission
		var policy jobs.ExecutionPolicy
		err := p.store.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
			var err error
			req, policy, err = p.executionAuthority(ctx, tx, lease)
			return err
		})
		if err != nil {
			return jobs.Completion{}, productionExecutionFailure(err)
		}
		if err = guard.Authorize(ctx); err != nil {
			return jobs.Completion{}, productionExecutionFailure(err)
		}
		image, err := a.images.prepare(ctx, lease.Owner, lease.Installation, lease.Asset)
		if err != nil {
			if scoped := p.sourceFailure(ctx, lease, err); scoped != nil {
				return jobs.Completion{}, scoped
			}
			return jobs.Completion{}, productionExecutionFailure(err)
		}
		defer image.Release()
		cfg := req.Configuration
		var guardError, accountingError, dispatchError error
		runner, err := analysis.New(analysis.Protocol{
			Encode: func(model, instruction, dataURL, format string) ([]byte, error) {
				return providerhttp.EncodeLimitedVisual(model, instruction, dataURL, format, policy.OutputField, cfg.Limits.OutputTokens, policy.MaxRequestBytes, policy.MaxImageBytes)
			},
			Parse: providerhttp.ParseVisual,
		}, func(ctx context.Context, input analysis.DispatchRequest) (analysis.DispatchReply, error) {
			reply, err := a.dispatcher.Dispatch(ctx, providers.DispatchRequest{OwnerID: input.OwnerID, ProfileID: input.ProfileID, Revision: input.Revision, Body: input.Body, Authorize: input.Authorize})
			if err != nil {
				dispatchError = productionProviderFailure(err)
				return analysis.DispatchReply{}, providerhttp.ClassifyVisualFailure(err)
			}
			accountingError = p.recordUsage(ctx, lease, providerhttp.ParseVisualUsage(reply.Body))
			if accountingError != nil {
				clear(reply.Body)
				return analysis.DispatchReply{}, analysis.ErrDenied
			}
			return analysis.DispatchReply{Body: reply.Body}, nil
		})
		if err != nil {
			return jobs.Completion{}, jobs.ErrStorage
		}
		request := analysis.Request{Owner: lease.Owner, Installation: lease.Installation, Asset: lease.Asset, ProfileID: cfg.ProfileID, Revision: cfg.Revision, Format: cfg.Format, AllowJSON: cfg.AllowJSON, Languages: cfg.Languages, PrimaryLanguage: cfg.PrimaryLanguage, Image: image, Guard: analysis.Guard{
			Authorize: func(ctx context.Context) error {
				err := guard.Authorize(ctx)
				if err != nil {
					guardError = err
				}
				return err
			},
			Reserve: func(ctx context.Context) error {
				err := guard.Reserve(ctx)
				if err != nil {
					guardError = err
				}
				return err
			},
		}}
		var result *analysis.Result
		if cfg.Mode == "context-assisted" {
			preparer := &aiContextPreparer{images: a.images}
			var contextReq aiContextRequest
			request.Context, contextReq, err = p.frozenContext(ctx, lease, cfg, image, preparer, request.Guard.Authorize)
			if err == nil {
				result, err = a.analyzeContextWith(ctx, request, preparer, contextReq, runner.RunContext)
			}
		} else {
			result, err = a.analyzeWith(ctx, request, runner.RunVisual)
		}
		if accountingError != nil {
			return jobs.Completion{}, productionExecutionFailure(accountingError)
		}
		if guardError != nil {
			return jobs.Completion{}, productionExecutionFailure(guardError)
		}
		if errors.Is(err, errAIContextChanged) {
			return jobs.Completion{}, jobs.Failure{Code: jobs.Permanent}
		}
		if scoped := p.sourceFailure(ctx, lease, err); scoped != nil {
			return jobs.Completion{}, scoped
		}
		if dispatchError != nil {
			return jobs.Completion{}, dispatchError
		}
		if err != nil {
			return jobs.Completion{}, productionExecutionFailure(err)
		}
		info := result.Info()
		return jobs.Completion{Proposal: result.Proposal(), SourceDigest: info.Image.Binding.SourceDigest, ImageDigest: info.Image.Digest, PromptVersion: info.PromptVersion, SchemaVersion: info.SchemaVersion}, nil
	}
}
