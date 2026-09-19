package capabilities_test

import (
	"testing"

	"immich-places-backend/internal/ai/capabilities"
)

func TestLegacyPermissiveProtocolCannotRemainApplicable(t *testing.T) {
	legacy := capabilities.Report{
		Revision: 1, ProtocolVersion: "capability-v1", PolicyFingerprint: "policy", Lifecycle: "completed",
		Observations: capabilities.Observations{Image: capabilities.Observation{Status: capabilities.StatusSupported}},
	}
	current := capabilities.ApplicabilityContext{
		AIEnabled: true, ProfileEnabled: true, ActiveRevision: 1,
		CurrentPolicyFingerprint: "policy",
	}
	if capabilities.IsApplicable(legacy, current) {
		t.Fatal("evidence from the permissive JSON/admission protocol requires an explicit retest")
	}
}
