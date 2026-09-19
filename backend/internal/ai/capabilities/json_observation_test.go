package capabilities_test

import (
	"testing"

	"immich-places-backend/internal/ai/capabilities"
)

func TestEvaluateJSONObservationSupported(t *testing.T) {
	fixture := capabilities.Fixture{Color: "blue", Shape: "circle"}
	obs := capabilities.EvaluateJSONObservation(fixture, `{"color":"blue","shape":"circle"}`)
	if obs.Status != capabilities.StatusSupported {
		t.Fatalf("status = %q reason=%q, want supported", obs.Status, obs.Reason)
	}
}

func TestEvaluateJSONObservationRejectsWrongFacts(t *testing.T) {
	fixture := capabilities.Fixture{Color: "blue", Shape: "circle"}
	obs := capabilities.EvaluateJSONObservation(fixture, `{"color":"red","shape":"square"}`)
	if obs.Status != capabilities.StatusUnverified || obs.Reason != "wrong_fixture_facts" {
		t.Fatalf("got status=%q reason=%q", obs.Status, obs.Reason)
	}
}

func TestEvaluateJSONObservationRejectsDuplicateKeys(t *testing.T) {
	fixture := capabilities.Fixture{Color: "blue", Shape: "circle"}
	obs := capabilities.EvaluateJSONObservation(fixture, `{"color":"blue","color":"red","shape":"circle"}`)
	if obs.Status != capabilities.StatusUnverified || obs.Reason != "invalid_output" {
		t.Fatalf("got status=%q reason=%q, want unverified invalid_output", obs.Status, obs.Reason)
	}
}

func TestEvaluateJSONObservationRejectsMalformedAndSchemaInvalid(t *testing.T) {
	fixture := capabilities.Fixture{Color: "blue", Shape: "circle"}
	cases := []string{
		`{color:blue}`,
		`{"color":"blue","shape":"circle","extra":true}`,
		`{"color":null,"shape":"circle"}`,
		"```json\n{\"color\":\"blue\",\"shape\":\"circle\"}\n```",
		`{"color":"blue","shape":"circle"}{"color":"blue","shape":"circle"}`,
	}
	for _, sample := range cases {
		obs := capabilities.EvaluateJSONObservation(fixture, sample)
		if obs.Status != capabilities.StatusUnverified || obs.Reason != "invalid_output" {
			t.Fatalf("sample %q => status=%q reason=%q", sample, obs.Status, obs.Reason)
		}
	}
}
