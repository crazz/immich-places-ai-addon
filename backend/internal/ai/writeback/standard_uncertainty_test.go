package writeback

import (
	"testing"

	"immich-places-backend/internal/ai/writepreview"
)

func TestStandardReadbackDoesNotSettleUnknownPartialOrUnavailableFields(t *testing.T) {
	zero, next := 0.0, 12.0
	op := Operation{Attempts: 1, Plan: writepreview.Plan{Version: "standard-preview-v2", ImageIdentity: "image", Fields: []string{"gps", "description"}, Intended: writepreview.Point{Latitude: zero, Longitude: next}, Description: &writepreview.DescriptionPlan{Before: writepreview.TextObservation{Presence: "null"}, Intended: "new"}}}
	fresh := writepreview.Metadata{ImageIdentity: "image", GPS: writepreview.GPS{Latitude: &zero, Longitude: &next}}
	for _, completed := range []bool{false, true} {
		decision := Readback(op, fresh, completed)
		if decision.Status != "verifying" || decision.Verified || !decision.GPSVerified || decision.Fields[1].Status != "unavailable" {
			t.Fatalf("unavailable field cannot settle partial: %+v", decision)
		}
	}
	fresh.Description = &writepreview.TextObservation{Presence: "absent"}
	if decision := Readback(op, fresh, false); decision.Status != "verifying" || decision.Code != "PARTIAL_UNRESOLVED" {
		t.Fatalf("unknown sender cannot settle partial: %+v", decision)
	}
	fresh.GPS = writepreview.GPS{}
	if decision := Readback(op, fresh, true); decision.Status != "retryable" {
		t.Fatal("all selected baseline with known completion must allow bounded retry", decision)
	}
	op.Attempts = 2
	if decision := Readback(op, fresh, true); decision.Status != "failed" {
		t.Fatal("third attempt permitted", decision)
	}
	fresh.Description = &writepreview.TextObservation{Presence: "value", Value: "external edit"}
	if decision := Readback(op, fresh, true); decision.Status != "conflict" {
		t.Fatal("third value did not conflict", decision)
	}
	fresh.ImageIdentity = "other"
	if decision := Readback(op, fresh, true); decision.Code != "SOURCE_CHANGED" {
		t.Fatal("source conflict missing", decision)
	}
}
