package capabilities_test

import (
	"testing"

	"immich-places-backend/internal/ai/capabilities"
)

func TestLimiterRejectsExcessGlobalStarts(t *testing.T) {
	limiter := capabilities.NewLimiter(2, 1)
	if err := limiter.Acquire("u1"); err != nil {
		t.Fatal(err)
	}
	if err := limiter.Acquire("u2"); err != nil {
		t.Fatal(err)
	}
	if err := limiter.Acquire("u3"); err != capabilities.ErrBusy {
		t.Fatalf("got %v, want ErrBusy", err)
	}
	limiter.Release("u1")
	if err := limiter.Acquire("u3"); err != nil {
		t.Fatal(err)
	}
}
