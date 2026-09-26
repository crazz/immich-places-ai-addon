package writeback

import (
	"testing"

	"immich-places-backend/internal/ai/writepreview"
)

func TestStandardReadbackSeparatesVerifiedFieldsAndPreventsPartialRetry(t *testing.T) {
	zero, next := 0.0, 12.0
	before := writepreview.TextObservation{Presence: "value", Value: "user\r\n"}
	op := Operation{Attempts: 1, Plan: writepreview.Plan{Version: "standard-preview-v2", ImageIdentity: "image", Fields: []string{"gps", "description"}, Before: writepreview.GPS{}, Intended: writepreview.Point{Latitude: zero, Longitude: next}, Description: &writepreview.DescriptionPlan{Before: before, Intended: "new exact\r\n"}}}
	fresh := writepreview.Metadata{ImageIdentity: "image", GPS: writepreview.GPS{Latitude: &zero, Longitude: &next}, Description: &before}
	decision := Readback(op, fresh, true)
	if decision.Status != "partial" || decision.Verified || !decision.GPSVerified || len(decision.Fields) != 2 || decision.Fields[0].Status != "verified" || decision.Fields[1].Status != "baseline" {
		t.Fatalf("partial standard request cannot retry or succeed: %+v", decision)
	}
	fresh.Description = &writepreview.TextObservation{Presence: "value", Value: op.Plan.Description.Intended}
	decision = Readback(op, fresh, true)
	if decision.Status != "succeeded" || !decision.Verified || len(decision.Fields) != 2 || decision.Fields[1].Status != "verified" {
		t.Fatalf("all exact selected fields must verify: %+v", decision)
	}
	op.Plan.Fields = []string{"description"}
	fresh.GPS = writepreview.GPS{}
	decision = Readback(op, fresh, true)
	if decision.Status != "succeeded" || decision.GPSVerified || len(decision.Fields) != 1 || decision.Fields[0].Field != "description" {
		t.Fatalf("unselected GPS has no outcome or catalog authority: %+v", decision)
	}
}
