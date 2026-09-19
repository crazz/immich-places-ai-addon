package providerhttp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"immich-places-backend/internal/ai/providers"
)

func limitFailure(requestID string) error {
	return providers.NewTransportFailure(providers.FailureLimit, requestID, 0, nil)
}

func upstreamFailure(requestID string, status int, body []byte) error {
	failure := providers.NewTransportFailure(providers.FailureUpstream, requestID, status, nil)
	failure.Code = allowlistedProviderCode(body)
	return failure
}

func allowlistedProviderCode(body []byte) string {
	var envelope struct {
		Error struct {
			Code  string `json:"code"`
			Param string `json:"param"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &envelope) != nil {
		return ""
	}
	if envelope.Error.Code == "unsupported_value" && envelope.Error.Param == "response_format" {
		return "unsupported_response_format"
	}
	switch envelope.Error.Code {
	case "invalid_api_key", "rate_limit_exceeded", "insufficient_quota", "context_length_exceeded", "model_not_found", "unsupported_response_format":
		return envelope.Error.Code
	default:
		return ""
	}
}

func mapRequestError(requestID string, err error) error {
	if errors.Is(err, providers.ErrRedirect) {
		return providers.ErrRedirect
	}
	switch {
	case errors.Is(err, context.Canceled):
		return providers.NewTransportFailure(providers.FailureCanceled, requestID, 0, err)
	case errors.Is(err, context.DeadlineExceeded):
		return providers.NewTransportFailure(providers.FailureTimeout, requestID, 0, err)
	default:
		return providers.NewTransportFailure(providers.FailureConnection, requestID, 0, err)
	}
}

func isResponseHeaderLimitError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "server response headers exceeded")
}
