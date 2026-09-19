package main

import (
	"errors"
	"net/http"

	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/providers"
)

func classifyCapabilityProviderFailure(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, providers.ErrPolicyDenied) {
		return capabilities.ErrPolicy
	}
	failure, ok := providers.AsTransportFailure(err)
	if !ok {
		return err
	}
	switch failure.Code {
	case "invalid_api_key":
		return capabilities.ErrAuthentication
	case "model_not_found":
		return capabilities.ErrUnavailableModel
	case "rate_limit_exceeded", "insufficient_quota":
		return capabilities.ErrRateLimit
	case "unsupported_response_format":
		return &capabilities.ModeUnsupportedError{}
	}
	switch failure.Category {
	case providers.FailureConnection, providers.FailureTimeout:
		return capabilities.ErrNetwork
	case providers.FailureCanceled:
		return err
	case providers.FailureUpstream:
		if failure.Status == http.StatusUnauthorized || failure.Status == http.StatusForbidden {
			return capabilities.ErrAuthentication
		}
		if failure.Status >= 500 {
			return capabilities.ErrServer
		}
		if failure.Status == http.StatusTooManyRequests {
			return capabilities.ErrRateLimit
		}
		return err
	default:
		return err
	}
}
