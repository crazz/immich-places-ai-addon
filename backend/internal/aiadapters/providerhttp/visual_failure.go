package providerhttp

import (
	"context"
	"errors"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/providers"
)

func ClassifyVisualFailure(err error) error {
	for _, safe := range []error{context.Canceled, context.DeadlineExceeded, analysis.ErrDenied, analysis.ErrBudget, analysis.ErrLimit} {
		if errors.Is(err, safe) {
			return safe
		}
	}
	for _, denied := range []error{providers.ErrDisabled, providers.ErrUnavailable, providers.ErrCredential, providers.ErrPolicyDenied, providers.ErrRedirect} {
		if errors.Is(err, denied) {
			return analysis.ErrDenied
		}
	}
	if failure, ok := providers.AsTransportFailure(err); ok {
		switch failure.Category {
		case providers.FailureCanceled:
			return context.Canceled
		case providers.FailureTimeout:
			return context.DeadlineExceeded
		case providers.FailureLimit:
			return analysis.ErrLimit
		case providers.FailureUpstream:
			if failure.Status == 401 || failure.Status == 403 {
				return analysis.ErrDenied
			}
			if (failure.Status == 400 || failure.Status == 422) && failure.Code == "unsupported_response_format" {
				return analysis.ErrUnsupportedFormat
			}
		}
	}
	return analysis.ErrUpstream
}
