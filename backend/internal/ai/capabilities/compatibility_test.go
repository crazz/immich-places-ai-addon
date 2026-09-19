package capabilities_test

import (
	"testing"

	"immich-places-backend/internal/ai/capabilities"
)

func TestCompatibilitySummaryStrictSchemaCompatible(t *testing.T) {
	obs := capabilities.Observations{
		Image:  capabilities.Observation{Status: capabilities.StatusSupported},
		JSON:   capabilities.Observation{Status: capabilities.StatusSupported},
		Strict: capabilities.Observation{Status: capabilities.StatusSupported},
	}
	if got := capabilities.CompatibilitySummary(obs); got != "strict-schema sample compatible" {
		t.Fatalf("compatibility = %q, want strict-schema sample compatible", got)
	}

	jsonUnsupported := obs
	jsonUnsupported.JSON = capabilities.Observation{Status: capabilities.StatusUnsupported, Reason: "unsupported_mode"}
	if got := capabilities.CompatibilitySummary(jsonUnsupported); got != "strict-schema sample compatible" {
		t.Fatalf("strict with JSON unsupported = %q, want strict-schema sample compatible", got)
	}
}

func TestCompatibilitySummaryJSONOnlyCompatible(t *testing.T) {
	obs := capabilities.Observations{
		Image:  capabilities.Observation{Status: capabilities.StatusSupported},
		JSON:   capabilities.Observation{Status: capabilities.StatusSupported},
		Strict: capabilities.Observation{Status: capabilities.StatusUnsupported, Reason: "unsupported_mode"},
	}
	if got := capabilities.CompatibilitySummary(obs); got != "json-only compatible" {
		t.Fatalf("compatibility = %q, want json-only compatible", got)
	}
}

func TestCompatibilitySummaryBothOutputModesUnsupported(t *testing.T) {
	obs := capabilities.Observations{
		Image:  capabilities.Observation{Status: capabilities.StatusSupported},
		JSON:   capabilities.Observation{Status: capabilities.StatusUnsupported, Reason: "unsupported_mode"},
		Strict: capabilities.Observation{Status: capabilities.StatusUnsupported, Reason: "unsupported_mode"},
	}
	if got := capabilities.CompatibilitySummary(obs); got != "unsupported" {
		t.Fatalf("compatibility = %q, want unsupported", got)
	}
}
