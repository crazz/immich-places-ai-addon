package capabilities

type ApplicabilityContext struct {
	AIEnabled                bool
	ProfileEnabled           bool
	ActiveRevision           int
	CurrentProtocolVersion   string
	CurrentPolicyFingerprint string
}

func IsApplicable(report Report, ctx ApplicabilityContext) bool {
	if report.Lifecycle != "completed" {
		return false
	}
	if ctx.CurrentProtocolVersion == "" {
		ctx.CurrentProtocolVersion = ProtocolVersion
	}
	if report.ProtocolVersion != ctx.CurrentProtocolVersion {
		return false
	}
	if report.PolicyFingerprint == "" || report.PolicyFingerprint != ctx.CurrentPolicyFingerprint {
		return false
	}
	if !ctx.AIEnabled || !ctx.ProfileEnabled {
		return false
	}
	if report.Revision != ctx.ActiveRevision {
		return false
	}
	return report.Observations.Image.Status == StatusSupported
}
