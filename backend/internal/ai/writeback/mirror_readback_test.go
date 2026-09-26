package writeback

import (
	"encoding/json"
	"testing"

	"immich-places-backend/internal/ai/writepreview"
)

func TestMirrorReadbackSeparatesObservedContentFromSenderSettlement(t *testing.T) {
	before := writepreview.MirrorBaseline{Present: true, Value: json.RawMessage(`{"x":0}`)}
	plan := writepreview.Plan{ImageIdentity: "source", Mirror: &writepreview.MirrorPlan{Before: before, Value: json.RawMessage(`{"x":1,"text":"exact"}`)}}
	for _, tc := range []struct {
		name, image, raw string
		known            bool
		attempts         int
		status, code     string
		verified         bool
	}{
		{"semantic match", "source", `{"text":"exact","x":1.0}`, true, 1, "succeeded", "METADATA_VERIFIED", true},
		{"unknown sender observed", "source", `{"x":1,"text":"exact"}`, false, 1, "verifying", "METADATA_OBSERVED_UNRESOLVED", true},
		{"known unchanged", "source", `{"x":0}`, true, 1, "retryable", "METADATA_RETRY_AVAILABLE", false},
		{"unknown unchanged", "source", `{"x":0}`, false, 1, "verifying", "RECONCILIATION_REQUIRED", false},
		{"exhausted", "source", `{"x":0}`, true, 2, "failed", "ATTEMPTS_EXHAUSTED", false},
		{"changed text", "source", `{"x":1,"text":"edited"}`, true, 1, "conflict", "METADATA_CONFLICT", false},
		{"changed image", "other", `{"x":1,"text":"exact"}`, true, 1, "conflict", "SOURCE_CHANGED", false},
		{"unavailable", "source", "", true, 1, "verifying", "METADATA_UNAVAILABLE", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fresh := writepreview.Metadata{ImageIdentity: tc.image}
			if tc.raw != "" {
				fresh.Mirror = &writepreview.MirrorBaseline{Present: true, Value: json.RawMessage(tc.raw)}
			}
			got := MirrorReadback(plan, fresh, tc.known, tc.attempts, false)
			if got.Status != tc.status || got.Code != tc.code || got.Verified != tc.verified {
				t.Fatalf("unexpected decision %+v", got)
			}
		})
	}
}
