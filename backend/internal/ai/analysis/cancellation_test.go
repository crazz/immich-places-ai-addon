package analysis_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestVisualDeadlineAndCancellationPreventLatePublication(t *testing.T) {
	for _, stage := range []string{"before", "dispatch", "after"} {
		t.Run(stage, func(t *testing.T) {
			runner, req, calls, _ := visualFixture(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if stage == "before" {
				cancel()
			}
			if stage == "dispatch" {
				req.Guard.Reserve = func(context.Context) error { cancel(); return nil }
			}
			if stage == "after" {
				checks := 0
				req.Guard.Authorize = func(context.Context) error {
					checks++
					if checks == 3 {
						cancel()
					}
					return nil
				}
			}
			result, err := runner.RunVisual(ctx, req)
			if !errors.Is(err, context.Canceled) || result != nil {
				t.Fatal("canceled result published", err)
			}
			if stage == "before" && *calls != 0 {
				t.Fatal("canceled input dispatched")
			}
		})
	}
	_, req, _, _ := visualFixture(t)
	for _, earlier := range []bool{false, true} {
		runner, err := analysis.New(analysis.Protocol{Encode: providerhttp.EncodeVisual, Parse: providerhttp.ParseVisual}, func(ctx context.Context, r analysis.DispatchRequest) (analysis.DispatchReply, error) {
			deadline, ok := ctx.Deadline()
			if !ok || time.Until(deadline) > 120*time.Second {
				t.Fatal("missing overall deadline")
			}
			if earlier && time.Until(deadline) > time.Second {
				t.Fatal("caller deadline extended")
			}
			return analysis.DispatchReply{}, context.DeadlineExceeded
		})
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		cancel := func() {}
		if earlier {
			ctx, cancel = context.WithTimeout(ctx, time.Second)
		}
		_, err = runner.RunVisual(ctx, req)
		cancel()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
	}
}
