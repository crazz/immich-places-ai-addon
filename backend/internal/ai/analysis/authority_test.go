package analysis_test

import (
	"context"
	"errors"
	"testing"

	"immich-places-backend/internal/ai/analysis"
)

func TestVisualReauthorizesBeforeEncodingTransmissionAndPublication(t *testing.T) {
	for _, rejectAt := range []int{1, 2, 3} {
		t.Run(string(rune('0'+rejectAt)), func(t *testing.T) {
			runner, req, calls, reservations := visualFixture(t)
			checks := 0
			req.Guard.Authorize = func(context.Context) error {
				checks++
				if checks == rejectAt {
					return errors.New("private denial marker")
				}
				return nil
			}
			result, err := runner.RunVisual(context.Background(), req)
			if !errors.Is(err, analysis.ErrDenied) || result != nil {
				t.Fatal("authority loss published", err)
			}
			want := 0
			if rejectAt == 3 {
				want = 1
			}
			if *calls != want || *reservations != want {
				t.Fatal("wrong dispatch count")
			}
		})
	}
}
