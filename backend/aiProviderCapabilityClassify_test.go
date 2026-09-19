package main

import (
	"errors"
	"net/http"
	"testing"

	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/providers"
)

func TestClassifyCapabilityProviderFailureMapsSafeCategories(t *testing.T) {
	cases := []struct {
		name     string
		failure  *providers.TransportFailure
		wantErr  error
		wantMode bool
	}{
		{
			name:    "authentication",
			failure: withCode(providers.NewTransportFailure(providers.FailureUpstream, "req", http.StatusUnauthorized, nil), "invalid_api_key"),
			wantErr: capabilities.ErrAuthentication,
		},
		{
			name:    "unavailable model",
			failure: withCode(providers.NewTransportFailure(providers.FailureUpstream, "req", http.StatusNotFound, nil), "model_not_found"),
			wantErr: capabilities.ErrUnavailableModel,
		},
		{
			name:    "rate limit",
			failure: withCode(providers.NewTransportFailure(providers.FailureUpstream, "req", http.StatusTooManyRequests, nil), "rate_limit_exceeded"),
			wantErr: capabilities.ErrRateLimit,
		},
		{
			name:    "server",
			failure: providers.NewTransportFailure(providers.FailureUpstream, "req", http.StatusBadGateway, nil),
			wantErr: capabilities.ErrServer,
		},
		{
			name:    "network",
			failure: providers.NewTransportFailure(providers.FailureConnection, "req", 0, nil),
			wantErr: capabilities.ErrNetwork,
		},
		{
			name:     "unsupported json mode",
			failure:  withCode(providers.NewTransportFailure(providers.FailureUpstream, "req", http.StatusBadRequest, nil), "unsupported_response_format"),
			wantErr:  capabilities.ErrModeUnsupported,
			wantMode: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyCapabilityProviderFailure(tc.failure)
			if !errors.Is(got, tc.wantErr) {
				t.Fatalf("got %v, want %v", got, tc.wantErr)
			}
			if tc.wantMode {
				var mode *capabilities.ModeUnsupportedError
				if !errors.As(got, &mode) {
					t.Fatal("expected ModeUnsupportedError")
				}
			}
		})
	}
}

func TestClassifyCapabilityProviderFailureKeepsUnknownUpstreamUnverified(t *testing.T) {
	failure := providers.NewTransportFailure(providers.FailureUpstream, "req", http.StatusBadRequest, nil)
	got := classifyCapabilityProviderFailure(failure)
	if errors.Is(got, capabilities.ErrServer) {
		t.Fatal("generic 400 must not be invented as server")
	}
	if errors.Is(got, capabilities.ErrModeUnsupported) {
		t.Fatal("generic 400 must not be invented as unsupported mode")
	}
	var mode *capabilities.ModeUnsupportedError
	if errors.As(got, &mode) {
		t.Fatal("generic 400 must not become ModeUnsupportedError")
	}
	for _, banned := range []error{
		capabilities.ErrAuthentication,
		capabilities.ErrUnavailableModel,
		capabilities.ErrRateLimit,
		capabilities.ErrPolicy,
		capabilities.ErrNetwork,
	} {
		if errors.Is(got, banned) {
			t.Fatalf("generic 400 must not invent category %v", banned)
		}
	}
	if _, ok := providers.AsTransportFailure(got); !ok {
		t.Fatalf("unknown upstream must remain the transport failure, got %T %v", got, got)
	}
}

func withCode(failure *providers.TransportFailure, code string) *providers.TransportFailure {
	failure.Code = code
	return failure
}
