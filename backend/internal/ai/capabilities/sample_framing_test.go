package capabilities_test

import (
	"testing"

	"immich-places-backend/internal/ai/capabilities"
)

func TestSyntheticSampleRejectsTrailingClosingDelimiter(t *testing.T) {
	for _, suffix := range []string{"}", "]", " } ", " ] "} {
		input := `{"color":"blue","shape":"circle"}` + suffix
		if _, _, err := capabilities.ParseSyntheticSample(input); err == nil {
			t.Errorf("accepted malformed JSON with trailing delimiter %q", suffix)
		}
	}
}

func TestSyntheticJSONDoesNotRepairSchemaEnumValues(t *testing.T) {
	fixture := capabilities.Fixture{Color: "blue", Shape: "circle"}
	for _, input := range []string{
		`{"color":"BLUE","shape":"circle"}`,
		`{"color":"blue","shape":"CIRCLE"}`,
	} {
		if got := capabilities.EvaluateJSONObservation(fixture, input); got.Status == capabilities.StatusSupported {
			t.Errorf("schema-invalid enum values were repaired into supported evidence: %s", input)
		}
	}
}
