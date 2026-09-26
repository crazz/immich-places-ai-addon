package main

import (
	"strings"

	"immich-places-backend/internal/ai/writepreview"
)

// Only static adapter SQL passes through this version selector. It never
// rewrites payloads or routes writes through views, preserving CAS row counts.
func aiWriteStatement(plan writepreview.Plan, statement string) string {
	switch plan.Version {
	case "gps-preview-v1":
		return statement
	case "standard-preview-v2":
		return strings.NewReplacer(
			"ai_write_operations", "ai_standard_write_operations",
			"ai_write_targets", "ai_standard_write_targets",
			"ai_write_events", "ai_standard_write_events",
			"ai_write_previews", "ai_standard_write_previews",
		).Replace(statement)
	}
	return ""
}

func (s *aiWriteStore) permitsPlan(plan writepreview.Plan) bool {
	if !s.available() {
		return false
	}
	if plan.Version == "gps-preview-v1" {
		return true
	}
	if plan.Version == "mirror-preview-v4" {
		return plan.Mirror != nil && plan.Manifest != nil && s.capabilities.Allows(plan.Installation, s.profile, "metadata") && (len(plan.Manifest.Targets) == 1 || s.capabilities.Allows(plan.Installation, s.profile, "stack_gps")) && (plan.Description == nil || s.capabilities.Allows(plan.Installation, s.profile, "description")) && plan.PolicyID == s.capabilities.Identity(plan.Installation, s.profile)
	}
	if plan.Version == "stack-preview-v3" {
		return plan.Manifest != nil && s.capabilities.Allows(plan.Installation, s.profile, "stack_gps") && (plan.Description == nil || s.capabilities.Allows(plan.Installation, s.profile, "description")) && plan.PolicyID == s.capabilities.Identity(plan.Installation, s.profile)
	}
	return plan.Version == "standard-preview-v2" && s.capabilities.Allows(plan.Installation, s.profile, "description") && plan.PolicyID == s.capabilities.Identity(plan.Installation, s.profile)
}
