package capabilities_test

import (
	"testing"

	"immich-places-backend/internal/ai/capabilities"
)

func TestIsApplicableRequiresDecisionFourContext(t *testing.T) {
	supported := capabilities.Report{
		Revision:          1,
		ProtocolVersion:   capabilities.ProtocolVersion,
		PolicyFingerprint: "fp",
		Lifecycle:         "completed",
		Observations: capabilities.Observations{
			Image: capabilities.Observation{Status: capabilities.StatusSupported},
		},
	}
	ctx := capabilities.ApplicabilityContext{
		AIEnabled:                true,
		ProfileEnabled:           true,
		ActiveRevision:           1,
		CurrentProtocolVersion:   capabilities.ProtocolVersion,
		CurrentPolicyFingerprint: "fp",
	}
	if !capabilities.IsApplicable(supported, ctx) {
		t.Fatal("supported current evidence should be applicable")
	}
	unverified := supported
	unverified.Observations.Image.Status = capabilities.StatusUnverified
	if capabilities.IsApplicable(unverified, ctx) {
		t.Fatal("unverified evidence must not be applicable")
	}
	stalePolicy := supported
	ctx.CurrentPolicyFingerprint = "other"
	if capabilities.IsApplicable(stalePolicy, ctx) {
		t.Fatal("policy mismatch must not be applicable")
	}
}

func TestInvalidateAfterEditsOrPolicyChanges(t *testing.T) {
	supported := capabilities.Report{
		Revision:          1,
		ProtocolVersion:   capabilities.ProtocolVersion,
		PolicyFingerprint: "fp",
		Lifecycle:         "completed",
		Observations: capabilities.Observations{
			Image: capabilities.Observation{Status: capabilities.StatusSupported},
		},
	}
	base := capabilities.ApplicabilityContext{
		AIEnabled:                true,
		ProfileEnabled:           true,
		ActiveRevision:           1,
		CurrentProtocolVersion:   capabilities.ProtocolVersion,
		CurrentPolicyFingerprint: "fp",
	}
	if !capabilities.IsApplicable(supported, base) {
		t.Fatal("baseline successful report should be applicable")
	}

	newRevision := base
	newRevision.ActiveRevision = 2
	if capabilities.IsApplicable(supported, newRevision) {
		t.Fatal("new revision must not treat prior report as current usable proof")
	}

	disabled := base
	disabled.ProfileEnabled = false
	if capabilities.IsApplicable(supported, disabled) {
		t.Fatal("disabled profile must not treat prior report as current usable proof")
	}

	aiOff := base
	aiOff.AIEnabled = false
	if capabilities.IsApplicable(supported, aiOff) {
		t.Fatal("disabled installation must not treat prior report as current usable proof")
	}

	policyChanged := base
	policyChanged.CurrentPolicyFingerprint = "new-fp"
	if capabilities.IsApplicable(supported, policyChanged) {
		t.Fatal("policy change must not treat prior report as current usable proof")
	}

	protocolChanged := base
	protocolChanged.CurrentProtocolVersion = "a-different-protocol"
	if capabilities.IsApplicable(supported, protocolChanged) {
		t.Fatal("protocol change must not treat prior report as current usable proof")
	}
}
