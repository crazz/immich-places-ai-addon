package main

import (
	"context"
	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/ai/translations"
	"immich-places-backend/internal/aiadapters/providerhttp"
	"time"
)

type aiTextTranslator struct{ dispatcher *providers.Dispatcher }

func (p aiTextTranslator) Translate(ctx context.Context, task translations.Task, authorize func(context.Context) error) (translations.Outcome, *analysis.Usage, error) {
	policy := task.Policy
	body, err := providerhttp.EncodeTranslation(policy.Binding.Model, task.Request.Basis, task.Language, "json", policy.OutputField, min(policy.MaxOutputTokens, 4096))
	if err != nil {
		return translations.Outcome{}, nil, translations.Failure("invalid_request")
	}
	defer clear(body)
	if len(body) > policy.MaxRequestBytes {
		return translations.Outcome{}, nil, translations.Failure("budget")
	}
	reply, err := p.dispatcher.Dispatch(ctx, providers.DispatchRequest{OwnerID: task.Owner, ProfileID: task.Request.ProfileID, Revision: task.Request.ProfileRevision, Body: body, Timeout: 120 * time.Second, Authorize: authorize})
	if err != nil {
		return translations.Outcome{}, nil, err
	}
	defer clear(reply.Body)
	var usage *analysis.Usage
	if parsed, parseErr := providerhttp.ParseVisual(reply.Body); parseErr == nil {
		usage = parsed.Usage
	}
	out, err := providerhttp.ParseTranslation(reply.Body, task.Language)
	if err != nil {
		return out, usage, translations.Failure("invalid_response")
	}
	return out, usage, nil
}
