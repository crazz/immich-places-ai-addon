package capabilities_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"immich-places-backend/internal/ai/capabilities"
)

func TestRunStopsWhenContextCanceledBetweenProbes(t *testing.T) {
	protocol := &recordingProtocol{}
	var hits atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	runner := &capabilities.Runner{
		Protocol: protocol,
		Select:   capabilities.FixedSelector(0),
		Clock:    fixedClock{now: time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)},
		Dispatch: func(ctx context.Context, ownerID, profileID string, revision int, body []byte) ([]byte, error) {
			n := hits.Add(1)
			if n == 1 {
				cancel()
				return []byte("TEXT:color: blue\nshape: circle"), nil
			}
			t.Fatalf("subsequent probe admitted after cancel, hit=%d", n)
			return nil, nil
		},
	}
	report, err := runner.Run(ctx, baseProbeRequest())
	if !errors.Is(err, context.Canceled) && report.Observations.JSON.Reason != "canceled" {
		t.Fatalf("expected cancel stop, err=%v report=%+v", err, report)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d, want 1", hits.Load())
	}
	if !report.InputMayBeConsumed {
		t.Fatal("already transmitted input must be described as potentially consumed")
	}
	if report.Observations.Strict.Status != capabilities.StatusUnverified {
		t.Fatalf("strict must remain unverified: %+v", report.Observations.Strict)
	}
}

func TestRevisionChangeBetweenProbesStopsAdmission(t *testing.T) {
	protocol := &recordingProtocol{}
	var hits atomic.Int32
	var admitHits atomic.Int32
	runner := &capabilities.Runner{
		Protocol: protocol,
		Select:   capabilities.FixedSelector(0),
		Clock:    fixedClock{now: time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)},
		Admit: func(ctx context.Context, ownerID, profileID string, revision int) error {
			n := admitHits.Add(1)
			if n >= 2 {
				return capabilities.ErrUnavailable
			}
			return nil
		},
		Dispatch: func(ctx context.Context, ownerID, profileID string, revision int, body []byte) ([]byte, error) {
			n := hits.Add(1)
			if n > 1 {
				t.Fatalf("subsequent probe admitted after revision change, hit=%d", n)
			}
			return []byte("TEXT:color: blue\nshape: circle"), nil
		},
	}
	report, err := runner.Run(context.Background(), baseProbeRequest())
	if !errors.Is(err, capabilities.ErrUnavailable) {
		t.Fatalf("expected revision-change stop, err=%v report=%+v", err, report)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d, want 1", hits.Load())
	}
	if admitHits.Load() < 2 {
		t.Fatalf("admit hits = %d, want at least 2", admitHits.Load())
	}
	if !report.InputMayBeConsumed {
		t.Fatal("already transmitted input must be described as potentially consumed")
	}
	if report.Observations.JSON.Status != capabilities.StatusUnverified || report.Observations.Strict.Status != capabilities.StatusUnverified {
		t.Fatalf("later probes must stay unverified: %+v", report.Observations)
	}
	if report.Lifecycle == "completed" && report.Applicable {
		t.Fatal("revision-change stop must not publish applicable success")
	}
}

func TestDeadlineExpiresBetweenProbesStopsAdmission(t *testing.T) {
	protocol := &recordingProtocol{}
	var hits atomic.Int32
	var expired atomic.Bool
	ctx := &errContext{errFn: func() error {
		if expired.Load() {
			return context.DeadlineExceeded
		}
		return nil
	}}
	runner := &capabilities.Runner{
		Protocol: protocol,
		Select:   capabilities.FixedSelector(0),
		Clock:    fixedClock{now: time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)},
		Dispatch: func(ctx context.Context, ownerID, profileID string, revision int, body []byte) ([]byte, error) {
			n := hits.Add(1)
			if n == 1 {
				expired.Store(true)
				return []byte("TEXT:color: blue\nshape: circle"), nil
			}
			t.Fatalf("subsequent probe admitted after deadline, hit=%d", n)
			return nil, nil
		},
	}
	report, err := runner.Run(ctx, baseProbeRequest())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline stop, err=%v report=%+v", err, report)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d, want 1", hits.Load())
	}
	if !report.InputMayBeConsumed {
		t.Fatal("already transmitted input must be described as potentially consumed")
	}
	if report.Observations.JSON.Reason != "timeout" {
		t.Fatalf("json reason = %q, want timeout", report.Observations.JSON.Reason)
	}
	if report.Lifecycle == "completed" {
		t.Fatal("deadline must not complete as success")
	}
}

type errContext struct {
	context.Context
	errFn func() error
}

func (c *errContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *errContext) Done() <-chan struct{}       { return nil }
func (c *errContext) Err() error                  { return c.errFn() }
func (c *errContext) Value(key any) any           { return nil }

func TestDisablementBetweenProbesStopsAdmission(t *testing.T) {
	protocol := &recordingProtocol{}
	var hits atomic.Int32
	var admitHits atomic.Int32
	runner := &capabilities.Runner{
		Protocol: protocol,
		Select:   capabilities.FixedSelector(0),
		Clock:    fixedClock{now: time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)},
		Admit: func(ctx context.Context, ownerID, profileID string, revision int) error {
			n := admitHits.Add(1)
			if n >= 2 {
				return capabilities.ErrDisabled
			}
			return nil
		},
		Dispatch: func(ctx context.Context, ownerID, profileID string, revision int, body []byte) ([]byte, error) {
			n := hits.Add(1)
			if n > 1 {
				t.Fatalf("subsequent probe admitted after disablement, hit=%d", n)
			}
			return []byte("TEXT:color: blue\nshape: circle"), nil
		},
	}
	report, err := runner.Run(context.Background(), baseProbeRequest())
	if !errors.Is(err, capabilities.ErrDisabled) {
		t.Fatalf("expected disablement stop, err=%v report=%+v", err, report)
	}
	if hits.Load() != 1 {
		t.Fatalf("hits = %d, want 1", hits.Load())
	}
	if !report.InputMayBeConsumed {
		t.Fatal("already transmitted input must be described as potentially consumed")
	}
	if report.Observations.JSON.Reason != "authority_revoked" {
		t.Fatalf("json reason = %q, want authority_revoked", report.Observations.JSON.Reason)
	}
}
