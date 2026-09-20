package main

import (
	"context"
	"errors"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func productionProviderFailure(err error) error {
	if failure, ok := providers.AsTransportFailure(err); ok && failure.Category == providers.FailureUpstream {
		switch failure.Code {
		case "invalid_api_key", "insufficient_quota", "model_not_found", "unsupported_response_format":
			return jobs.Failure{Code: jobs.Blocked}
		}
		switch failure.Status {
		case 401, 402, 403, 404:
			return jobs.Failure{Code: jobs.Blocked}
		case 408, 425, 429:
			return jobs.Failure{Code: jobs.Transient}
		}
		if failure.Status >= 500 {
			return jobs.Failure{Code: jobs.Transient}
		}
		return jobs.Failure{Code: jobs.Permanent}
	}
	return productionExecutionFailure(providerhttp.ClassifyVisualFailure(err))
}
func productionExecutionFailure(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, jobs.ErrStorage) || errors.Is(err, jobs.ErrLease) || errors.Is(err, jobs.ErrBudget) {
		return err
	}
	var imageFailure aiImageHTTPError
	if errors.As(err, &imageFailure) {
		if imageFailure.Status == 401 {
			return jobs.Failure{Code: jobs.Blocked}
		}
		if imageFailure.Status < 500 && imageFailure.Status != 408 && imageFailure.Status != 429 {
			return jobs.Failure{Code: jobs.Permanent}
		}
	}
	if errors.Is(err, jobs.ErrDenied) || errors.Is(err, analysis.ErrDenied) || errors.Is(err, analysis.ErrUnsupportedFormat) {
		return jobs.Failure{Code: jobs.Blocked}
	}
	if errors.Is(err, analysis.ErrUpstream) || errors.Is(err, errAIImageUpstream) {
		return jobs.Failure{Code: jobs.Transient}
	}
	return err
}
