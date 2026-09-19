package capabilities_test

import (
	"errors"
	"testing"

	"immich-places-backend/internal/ai/capabilities"
)

func TestClassifyParseFailureDistinguishesToolFromRefusal(t *testing.T) {
	if got := capabilities.ClassifyParseFailure(capabilities.ErrToolResponse); got != "tool_response" {
		t.Fatalf("tool err => %q, want tool_response", got)
	}
	if got := capabilities.ClassifyParseFailure(capabilities.ErrRefusal); got != "refusal" {
		t.Fatalf("refusal err => %q, want refusal", got)
	}
	if got := capabilities.ClassifyParseFailure(capabilities.ErrTruncation); got != "truncation" {
		t.Fatalf("truncation err => %q, want truncation", got)
	}
	legacy := errors.New("capability response refused or requested tools")
	if got := capabilities.ClassifyParseFailure(legacy); got != "invalid_output" {
		t.Fatalf("untyped combined message => %q, want invalid_output (not refusal via substring)", got)
	}
}
