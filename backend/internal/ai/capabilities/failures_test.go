package capabilities_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"immich-places-backend/internal/ai/capabilities"
)

func TestRunDistinguishesAuthFailureFromUnsupportedMode(t *testing.T) {
	protocol := &recordingProtocol{}
	var hits atomic.Int32
	runner := &capabilities.Runner{
		Protocol: protocol,
		Select:   capabilities.FixedSelector(0),
		Clock:    fixedClock{now: time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)},
		Dispatch: func(ctx context.Context, ownerID, profileID string, revision int, body []byte) ([]byte, error) {
			n := hits.Add(1)
			switch n {
			case 1:
				return []byte("TEXT:color: blue\nshape: circle"), nil
			case 2:
				return nil, capabilities.ErrAuthentication
			default:
				t.Fatalf("remaining probes must stop after auth failure, hit=%d", n)
				return nil, nil
			}
		},
	}
	report, err := runner.Run(context.Background(), baseProbeRequest())
	if err != nil {
		t.Fatalf("auth failure should persist as report, err=%v", err)
	}
	if hits.Load() != 2 {
		t.Fatalf("hits = %d, want 2", hits.Load())
	}
	if report.Observations.Image.Status != capabilities.StatusSupported {
		t.Fatalf("image = %+v", report.Observations.Image)
	}
	if report.Observations.JSON.Status != capabilities.StatusUnverified || report.Observations.JSON.Reason != "authentication" {
		t.Fatalf("json = %+v", report.Observations.JSON)
	}
	if report.Observations.Strict.Status != capabilities.StatusUnverified {
		t.Fatalf("strict must remain unverified, got %+v", report.Observations.Strict)
	}
	if report.Compatibility != "failed" {
		t.Fatalf("compatibility = %q, want failed", report.Compatibility)
	}
	if report.Observations.JSON.Reason == "unsupported_mode" || report.Observations.Strict.Status == capabilities.StatusSupported {
		t.Fatal("must not reinterpret auth failure as unsupported mode or JSON fallback success")
	}
}

func TestRunStopsOnInvalidOutputWithoutRepair(t *testing.T) {
	protocol := &recordingProtocol{}
	var hits atomic.Int32
	runner := &capabilities.Runner{
		Protocol: protocol,
		Select:   capabilities.FixedSelector(0),
		Clock:    fixedClock{now: time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)},
		Dispatch: func(ctx context.Context, ownerID, profileID string, revision int, body []byte) ([]byte, error) {
			n := hits.Add(1)
			switch n {
			case 1:
				return []byte("TEXT:color: blue\nshape: circle"), nil
			case 2:
				return []byte(`JSON:{"color":"red","shape":"square"}`), nil
			default:
				t.Fatalf("must not repair or continue after wrong JSON facts, hit=%d", n)
				return nil, nil
			}
		},
	}
	report, err := runner.Run(context.Background(), baseProbeRequest())
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 {
		t.Fatalf("hits = %d, want 2", hits.Load())
	}
	if report.Observations.JSON.Status != capabilities.StatusUnverified || report.Observations.JSON.Reason != "wrong_fixture_facts" {
		t.Fatalf("json = %+v", report.Observations.JSON)
	}
	if report.Observations.Strict.Status != capabilities.StatusUnverified {
		t.Fatalf("strict must stay unverified, got %+v", report.Observations.Strict)
	}
	if report.Compatibility == "json-only compatible" || report.Compatibility == "strict-schema sample compatible" {
		t.Fatalf("invalid sample must not produce compatibility pass: %q", report.Compatibility)
	}
}
