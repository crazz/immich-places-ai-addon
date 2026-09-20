package analysis_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestContextResourceAndReservationFailuresNeverPublish(t *testing.T) {
	for _, failure := range []string{"request", "response", "reservation", "double-reservation", "unreserved", "cancel-before", "cancel-after", "deadline"} {
		t.Run(failure, func(t *testing.T) {
			_, req, _, reservations := visualFixture(t)
			req.Context = contextBundle(t, req)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if failure == "cancel-before" {
				cancel()
			}
			if failure == "reservation" {
				req.Guard.Reserve = func(context.Context) error { *reservations++; return errors.New("private budget") }
			}
			protocol := analysis.Protocol{Encode: providerhttp.EncodeVisual, Parse: providerhttp.ParseVisual}
			if failure == "request" {
				protocol.Encode = func(_, _, _, _ string) ([]byte, error) { return make([]byte, (15<<20)+1), nil }
			}
			content, err := os.ReadFile("../results/testdata/ai-analysis-result.unknown.json")
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			runner, err := analysis.New(protocol, func(ctx context.Context, in analysis.DispatchRequest) (analysis.DispatchReply, error) {
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) > 120*time.Second {
					t.Fatal("unbounded context attempt")
				}
				if failure == "deadline" {
					return analysis.DispatchReply{}, context.DeadlineExceeded
				}
				if failure != "unreserved" {
					if err := in.Authorize(ctx); err != nil {
						return analysis.DispatchReply{}, err
					}
				}
				if failure == "double-reservation" {
					return analysis.DispatchReply{}, in.Authorize(ctx)
				}
				calls++
				if failure == "cancel-after" {
					cancel()
				}
				if failure == "response" {
					return analysis.DispatchReply{Body: make([]byte, (1<<20)+1)}, nil
				}
				return analysis.DispatchReply{Body: content}, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			result, err := runner.RunContext(ctx, req)
			if result != nil || err == nil || calls > 1 || *reservations > 1 {
				t.Fatal("partial success or hidden retry", err, calls, *reservations)
			}
			want := analysis.ErrBudget
			switch failure {
			case "request", "response":
				want = analysis.ErrLimit
			case "cancel-before", "cancel-after":
				want = context.Canceled
			case "deadline":
				want = context.DeadlineExceeded
			}
			if !errors.Is(err, want) {
				t.Fatal("wrong failure boundary", err, want)
			}
			if (failure == "request" || failure == "cancel-before" || failure == "reservation") && calls != 0 {
				t.Fatal("unadmitted provider call")
			}
		})
	}
}
