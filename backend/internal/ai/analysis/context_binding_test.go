package analysis_test

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/contextual"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestContextBindingMismatchFailsBeforeEncoding(t *testing.T) {
	for _, field := range []string{"owner", "installation", "asset", "profile", "revision", "source"} {
		t.Run(field, func(t *testing.T) {
			_, req, _, reservations := visualFixture(t)
			binding := contextBundle(t, req).Info().Binding
			switch field {
			case "owner":
				binding.Owner = "other"
			case "installation":
				binding.Installation = "other"
			case "asset":
				binding.Asset = "other"
			case "profile":
				binding.Profile = "other"
			case "revision":
				binding.Revision++
			case "source":
				binding.SourceDigest = "other"
			}
			var err error
			req.Context, err = contextual.Build(contextual.Input{Binding: binding, Consent: contextual.Consent{Binding: binding, Version: contextual.Version}, Window: time.Hour})
			if err != nil {
				t.Fatal(err)
			}
			encoded := 0
			runner, err := analysis.New(analysis.Protocol{Encode: func(_, _, _, _ string) ([]byte, error) { encoded++; return []byte("body"), nil }, Parse: providerhttp.ParseVisual}, func(context.Context, analysis.DispatchRequest) (analysis.DispatchReply, error) {
				return analysis.DispatchReply{}, analysis.ErrUpstream
			})
			if err != nil {
				t.Fatal(err)
			}
			result, err := runner.RunContext(context.Background(), req)
			if err == nil || result != nil || encoded != 0 || *reservations != 0 {
				t.Fatal("mismatched bundle reached serialization", encoded, err)
			}
		})
	}
}
