package analysis_test

import (
	"context"
	"errors"
	"testing"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestVisualReservationIsSingleUseAndFailureIsSafe(t *testing.T) {
	for _, exhausted := range []bool{false, true} {
		_, req, _, reservations := visualFixture(t)
		if exhausted {
			req.Guard.Reserve = func(context.Context) error { *reservations++; return errors.New("private budget state") }
		}
		runner, err := analysis.New(analysis.Protocol{Encode: providerhttp.EncodeVisual, Parse: providerhttp.ParseVisual}, func(ctx context.Context, r analysis.DispatchRequest) (analysis.DispatchReply, error) {
			first := r.Authorize(ctx)
			if exhausted {
				return analysis.DispatchReply{}, first
			}
			if first != nil {
				t.Fatal(first)
			}
			return analysis.DispatchReply{}, r.Authorize(ctx)
		})
		if err != nil {
			t.Fatal(err)
		}
		result, err := runner.RunVisual(context.Background(), req)
		if !errors.Is(err, analysis.ErrBudget) || result != nil || *reservations != 1 {
			t.Fatal("reservation was reused or leaked", err, *reservations)
		}
	}
}

func TestVisualCannotPublishWithoutSuccessfulReservation(t *testing.T) {
	for _, ignoredFailure := range []bool{false, true} {
		_, req, _, _ := visualFixture(t)
		req.Guard.Reserve = func(context.Context) error { return errors.New("budget exhausted") }
		runner, _ := analysis.New(analysis.Protocol{Encode: providerhttp.EncodeVisual, Parse: providerhttp.ParseVisual}, func(ctx context.Context, r analysis.DispatchRequest) (analysis.DispatchReply, error) {
			if ignoredFailure {
				_ = r.Authorize(ctx)
			}
			return analysis.DispatchReply{Body: []byte(`{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":"{}"}}]}`)}, nil
		})
		result, err := runner.RunVisual(context.Background(), req)
		if result != nil || !errors.Is(err, analysis.ErrBudget) {
			t.Fatal("unreserved response accepted by workflow", err)
		}
	}
}
